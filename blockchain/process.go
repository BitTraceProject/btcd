// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"time"

	"github.com/BitTraceProject/BitTrace-Types/pkg/common"
	"github.com/BitTraceProject/BitTrace-Types/pkg/constants"
	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
	"github.com/btcsuite/btcd/bittrace"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/database"
)

// BehaviorFlags is a bitmask defining tweaks to the normal behavior when
// performing chain processing and consensus rules checks.
type BehaviorFlags uint32

const (
	// BFFastAdd may be set to indicate that several checks can be avoided
	// for the block since it is already known to fit into the chain due to
	// already proving it correct links into the chain up to a known
	// checkpoint.  This is primarily used for headers-first mode.
	BFFastAdd BehaviorFlags = 1 << iota

	// BFNoPoWCheck may be set to indicate the proof of work check which
	// ensures a block hashes to a value less than the required target will
	// not be performed.
	BFNoPoWCheck

	// BFNone is a convenience value to specifically indicate no flags.
	BFNone BehaviorFlags = 0
)

// blockExists determines whether a block with the given hash exists either in
// the main chain or any side chains.
//
// This function is safe for concurrent access.
func (b *BlockChain) blockExists(hash *chainhash.Hash) (bool, error) {
	// Check block index first (could be main chain or side chain blocks).
	if b.index.HaveBlock(hash) {
		return true, nil
	}

	// Check in the database.
	var exists bool
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		exists, err = dbTx.HasBlock(hash)
		if err != nil || !exists {
			return err
		}

		// Ignore side chain blocks in the database.  This is necessary
		// because there is not currently any record of the associated
		// block index data such as its block height, so it's not yet
		// possible to efficiently load the block and do anything useful
		// with it.
		//
		// Ultimately the entire block index should be serialized
		// instead of only the current main chain so it can be consulted
		// directly.
		_, err = dbFetchHeightByHash(dbTx, hash)
		if isNotInMainChainErr(err) {
			exists = false
			return nil
		}
		return err
	})
	return exists, err
}

// ZJH orphan process
// processOrphans determines if there are any orphans which depend on the passed
// block hash (they are no longer orphans if true) and potentially accepts them.
// It repeats the process for the newly accepted blocks (to detect further
// orphans which may no longer be orphans) until there are no more.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to maybeAcceptBlock.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) processOrphans(hash *chainhash.Hash, flags BehaviorFlags, traceData *bittrace.TraceData) error {
	var orphanProcessRevision = structure.NewRevision(structure.RevisionTypeOrphanProcess, traceData.GetSnapshotID(), structure.RevisionDataOrphanProcessInit{})

	// Start with processing at least the passed hash.  Leave a little room
	// for additional orphan blocks that need to be processed without
	// needing to grow the array in the common case.
	processHashes := make([]*chainhash.Hash, 0, 10)
	processHashes = append(processHashes, hash)
	for len(processHashes) > 0 {
		// Pop the first hash to process from the slice.
		processHash := processHashes[0]
		processHashes[0] = nil // Prevent GC leak.
		processHashes = processHashes[1:]

		// Look up all orphans that are parented by the block we just
		// accepted.  This will typically only be one, but it could
		// be multiple if multiple blocks are mined and broadcast
		// around the same time.  The one with the most proof of work
		// will eventually win out.  An indexing for loop is
		// intentionally used over a range here as range does not
		// reevaluate the slice on each iteration nor does it adjust the
		// index for the modified slice.
		for i := 0; i < len(b.prevOrphans[*processHash]); i++ {
			orphan := b.prevOrphans[*processHash][i]
			if orphan == nil {
				log.Warnf("Found a nil entry at index %d in the "+
					"orphan dependency list for block %v", i,
					processHash)
				continue
			}

			// Remove the orphan from the orphan pool.
			orphanHash := orphan.block.Hash()
			b.removeOrphanBlock(orphan)
			i--

			// Potentially accept the block into the block chain.
			isMainChain, err := b.maybeAcceptBlock(orphan.block, flags, traceData)
			if err != nil {
				return err
			}

			// orphan block 上了链，可能是 mainchain，也可能是 sidechain
			traceData.CommitEventOrphan(structure.EventTypeOrphanConnect, orphan.block.MsgBlock().Header.PrevBlock.String(), orphan.block.Hash().String(), isMainChain)
			// Add this block to the list of blocks to process so
			// any orphan blocks that depend on this block are
			// handled too.
			processHashes = append(processHashes, orphanHash)
		}
	}

	traceData.CommitRevision(orphanProcessRevision, time.Now(), structure.RevisionDataOrphanProcessFinal{})
	return nil
}

// ZJH block verify
// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block chain.  It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and insertion into
// the block chain along with best chain selection and reorganization.
//
// When no errors occurred during processing, the first return value indicates
// whether or not the block is on the main chain and the second indicates
// whether or not the block is an orphan.
//
// This function is safe for concurrent access.
func (b *BlockChain) ProcessBlock(block *btcutil.Block, flags BehaviorFlags, traceData *bittrace.TraceData, fromPeer string) (bool, bool, error) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	var (
		existInChain  bool
		existInOrphan bool
		isOrphan      bool
		err           error
	)

	fastAdd := flags&BFFastAdd == BFFastAdd

	blockHash := block.Hash()
	log.Tracef("Processing block %v", blockHash)

	_, existInOrphan = b.orphans[*blockHash]

	// generate init snapshot
	{
		var (
			forkHeight        int32 // 通过回溯 prevNode 找它
			targetChainID     string
			targetChainHeight int32 // 通过回溯 prevNode 计算
			initTime          = time.Now()
		)
		// The height of this block is one more than the referenced previous
		// block.
		prevHash := &block.MsgBlock().Header.PrevBlock
		prevNode := b.index.LookupNode(prevHash)

		if prevNode != nil {
			if prevNode.hash.IsEqual(&b.BestSnapshot().Hash) {
				// 1 如果该区块的父亲是主链的最优
				// 如果当前区块的前一个区块是 mainchain 的最优区块，则 forkHeight 直接赋值 0
				forkHeight = 0
				targetChainHeight = b.BestSnapshot().Height + 1
			} else {
				// 2 如果该区块的父亲不是主链的最优，继续判断是否处于一个 fork 上
				if b.index.NodeStatus(prevNode).KnownInvalid() {
					// 2.1 如果父亲状态未知，则他是孤儿
					isOrphan = true
				} else {
					b.bestChain.mtx.Lock()
					forkNode := b.bestChain.findFork(prevNode)
					if forkNode != nil {
						// 2.2 如果找得到 fork，处于侧链
						forkHeight = forkNode.height
						targetChainHeight = prevNode.height + 1
					} else {
						// 2.3 如果父亲不是主链最优且找不到其父亲，则父亲是孤儿区块，则它也是孤儿区块
						isOrphan = true
					}
					b.bestChain.mtx.Unlock()
				}
			}
		} else {
			// 2 如果该区块已位于孤儿，则是孤儿区块
			isOrphan = true
			if isOrphan {
				forkHeight = -1
				targetChainHeight = block.Height()
			}
		}
		if forkHeight == -1 {
			targetChainID = constants.ORPHAN_CHAIN_ID
		} else {
			targetChainID = common.GenChainID(forkHeight)
		}
		bestState := b.BestSnapshot()
		state := &structure.BestState{
			Hash:            bestState.Hash.String(),
			Height:          bestState.Height,
			Bits:            bestState.Bits,
			BlockSize:       bestState.BlockSize,
			BlockWeight:     bestState.BlockWeight,
			NumTxns:         bestState.NumTxns,
			TotalTxns:       bestState.TotalTxns,
			MedianTimestamp: common.FromTime(bestState.MedianTime),
		}
		if err := traceData.SetInitSnapshot(targetChainID, targetChainHeight, initTime, block.Hash().String(), state); err != nil {
			bittrace.Error("%v", err)
		}

		var receiveBlockRevision = structure.NewRevision(structure.RevisionTypeBlockReceive, traceData.GetSnapshotID(), structure.RevisionDataBlockReceiveInit{
			PeerIPAddr:      fromPeer,
			MinerWalletAddr: GetMinerAddress(block.MsgBlock(), b.chainParams),
		})
		// no status change, so no event and result
		traceData.CommitRevision(receiveBlockRevision, time.Now(), structure.RevisionDataBlockReceiveFinal{OK: true})
	}
	// 所有中途 return 都需要处理 revision

	var blockVerifyRevision = structure.NewRevision(structure.RevisionTypeBlockVerify, traceData.GetSnapshotID(), structure.RevisionDataBlockVerifyInit{
		Hash:           block.Hash().String(),
		ParentHash:     block.MsgBlock().Header.PrevBlock.String(),
		Height:         block.Height(),
		NumTxns:        uint64(len(block.Transactions())),
		Version:        block.MsgBlock().Header.Version,
		Bits:           block.MsgBlock().Header.Bits,
		WorkSum:        CalcWork(block.MsgBlock().Header.Bits).String(),
		MerkleRootHash: block.MsgBlock().Header.MerkleRoot.String(),
		Nonce:          block.MsgBlock().Header.Nonce,
	})
	// verify 1 是否已存在
	// The block must not already exist in the main chain or side chains.
	existInChain, err = b.blockExists(blockHash)
	if err != nil {
		return false, false, err
	}
	if existInChain {
		str := fmt.Sprintf("already have block %v", blockHash)
		traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
			OK:     false,
			Result: str,
		})
		return false, false, ruleError(ErrDuplicateBlock, str)
	}
	// verify 2 是否已是孤儿
	// The block must not already exist as an orphan.
	if existInOrphan {
		str := fmt.Sprintf("already have block (orphan) %v", blockHash)
		traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
			OK:     false,
			Result: str,
		})
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// verify 3
	// Perform preliminary sanity checks on the block and its transactions.
	err = checkBlockSanity(block, b.chainParams.PowLimit, b.timeSource, flags)
	if err != nil {
		traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
			OK:     false,
			Result: err.Error(),
		})
		return false, false, err
	}

	// Find the previous checkpoint and perform some additional checks based
	// on the checkpoint.  This provides a few nice properties such as
	// preventing old side chain blocks before the last checkpoint,
	// rejecting easy to mine, but otherwise bogus, blocks that could be
	// used to eat memory, and ensuring expected (versus claimed) proof of
	// work requirements since the previous checkpoint are met.
	blockHeader := &block.MsgBlock().Header
	checkpointNode, err := b.findPreviousCheckpoint()
	if err != nil {
		traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
			OK:     false,
			Result: err.Error(),
		})
		return false, false, err
	}
	if checkpointNode != nil {
		// Ensure the block timestamp is after the checkpoint timestamp.
		checkpointTime := time.Unix(checkpointNode.timestamp, 0)
		if blockHeader.Timestamp.Before(checkpointTime) {
			str := fmt.Sprintf("block %v has timestamp %v before "+
				"last checkpoint timestamp %v", blockHash,
				blockHeader.Timestamp, checkpointTime)
			traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
				OK:     false,
				Result: str,
			})
			return false, false, ruleError(ErrCheckpointTimeTooOld, str)
		}
		if !fastAdd {
			// Even though the checks prior to now have already ensured the
			// proof of work exceeds the claimed amount, the claimed amount
			// is a field in the block header which could be forged.  This
			// check ensures the proof of work is at least the minimum
			// expected based on elapsed time since the last checkpoint and
			// maximum adjustment allowed by the retarget rules.
			duration := blockHeader.Timestamp.Sub(checkpointTime)
			requiredTarget := CompactToBig(b.calcEasiestDifficulty(
				checkpointNode.bits, duration))
			currentTarget := CompactToBig(blockHeader.Bits)
			if currentTarget.Cmp(requiredTarget) > 0 {
				str := fmt.Sprintf("block target difficulty of %064x "+
					"is too low when compared to the previous "+
					"checkpoint", currentTarget)
				traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
					OK:     false,
					Result: str,
				})
				return false, false, ruleError(ErrDifficultyTooLow, str)
			}
		}
	}
	traceData.CommitRevision(blockVerifyRevision, time.Now(), structure.RevisionDataBlockVerifyFinal{
		OK: true,
	})

	// Handle orphan blocks.
	prevHash := &blockHeader.PrevBlock
	prevHashExists, err := b.blockExists(prevHash)
	if err != nil {
		return false, false, err
	}
	if !prevHashExists {
		log.Infof("Adding orphan block %v with parent %v", blockHash, prevHash)
		b.addOrphanBlock(block, traceData)

		return false, true, nil
	}

	// The block has passed all context independent checks and appears sane
	// enough to potentially accept it into the block chain.
	isMainChain, err := b.maybeAcceptBlock(block, flags, traceData)
	if err != nil {
		return false, false, err
	}

	// Accept any orphan blocks that depend on this block (they are
	// no longer orphans) and repeat for those accepted blocks until
	// there are no more.
	err = b.processOrphans(blockHash, flags, traceData)
	if err != nil {
		return false, false, err
	}

	log.Debugf("Accepted block %v", blockHash)

	return isMainChain, false, nil
}
