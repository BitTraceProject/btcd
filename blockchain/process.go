// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"time"

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
	var orphanProcessRevision = structure.NewRevision(structure.FromString("revision_orphan_process"), traceData.Snapshot.ID)

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
			_, err := b.maybeAcceptBlock(orphan.block, flags, traceData)
			if err != nil {
				return err
			}

			// Add this block to the list of blocks to process so
			// any orphan blocks that depend on this block are
			// handled too.
			processHashes = append(processHashes, orphanHash)
		}
	}

	if err := traceData.CommitRevision(orphanProcessRevision, fmt.Sprintf("processHashNum=%d", len(processHashes)), time.Now()); err != nil {
		bittrace.Error("%v", err)
	}
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
func (b *BlockChain) ProcessBlock(block *btcutil.Block, flags BehaviorFlags, traceData *bittrace.TraceData) (bool, bool, error) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	{
		// init snapshot
		var (
			forkHeight        int32 // 通过回溯 prevNode 找它
			targetChainID     = structure.GenChainID(forkHeight)
			targetChainHeight int32 // 通过回溯 prevNode 计算
			initTime          = time.Now()
			initStatus        = bittrace.GetFinalStatus()
		)
		// 搞成工具函数
		// The height of this block is one more than the referenced previous
		// block.
		prevHash := &block.MsgBlock().Header.PrevBlock
		prevNode := b.index.LookupNode(prevHash)

		if prevNode.hash.IsEqual(&b.bestChain.Tip().hash) {
			// 如果当前区块的前一个区块是 mainchain 的最优区块，则 forkHeight 直接赋值 0
			forkHeight = 0
			targetChainHeight = b.bestChain.Height() + 1
		} else {
			// 否则当前区块可能属于其他 sidechain，则 forkHeight 通过回溯计算
			// TODO 其他复杂情况的单独处理，如 prevBlock 不存在，prevBlock 是已知但是无效的
			if prevNode == nil {
				//str := fmt.Sprintf("previous block %s is unknown", prevHash)
				//return false, ruleError(ErrPreviousBlockUnknown, str)
			} else if b.index.NodeStatus(prevNode).KnownInvalid() {
				//str := fmt.Sprintf("previous block %s is known to be invalid", prevHash)
				//return false, ruleError(ErrInvalidAncestorBlock, str)
			} else {
				b.bestChain.mtx.Lock()
				forkHeight = b.bestChain.findFork(prevNode).height
				b.bestChain.mtx.Unlock()

				targetChainHeight = prevNode.height + 1
			}
		}
		initSnapshot := bittrace.InitSnapshot(targetChainID, targetChainHeight, initTime, initStatus)
		if err := traceData.SetInitSnapshot(&initSnapshot); err != nil {
			bittrace.Error("%v", err)
		}

		var receiveBlockRevision = structure.NewRevision(structure.FromString("revision_receive_block"), initSnapshot.ID)
		if err := traceData.CommitRevision(receiveBlockRevision, "receive_block", time.Now()); err != nil {
			bittrace.Error("%v", err)
		}
	}

	var blockVerifyRevision = structure.NewRevision(structure.FromString("revision_block_verify"), traceData.Snapshot.ID)

	fastAdd := flags&BFFastAdd == BFFastAdd

	blockHash := block.Hash()
	log.Tracef("Processing block %v", blockHash)

	// The block must not already exist in the main chain or side chains.
	exists, err := b.blockExists(blockHash)
	if err != nil {
		return false, false, err
	}
	if exists {
		str := fmt.Sprintf("already have block %v", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// The block must not already exist as an orphan.
	if _, exists := b.orphans[*blockHash]; exists {
		str := fmt.Sprintf("already have block (orphan) %v", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// Perform preliminary sanity checks on the block and its transactions.
	err = checkBlockSanity(block, b.chainParams.PowLimit, b.timeSource, flags)
	if err != nil {
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
		return false, false, err
	}
	if checkpointNode != nil {
		// Ensure the block timestamp is after the checkpoint timestamp.
		checkpointTime := time.Unix(checkpointNode.timestamp, 0)
		if blockHeader.Timestamp.Before(checkpointTime) {
			str := fmt.Sprintf("block %v has timestamp %v before "+
				"last checkpoint timestamp %v", blockHash,
				blockHeader.Timestamp, checkpointTime)

			if err := traceData.CommitRevision(blockVerifyRevision, "block_verify_failed timestamp before last checkpoint", time.Now()); err != nil {
				bittrace.Error("%v", err)
			}

			return false, false, ruleError(ErrCheckpointTimeTooOld, str)
		}
		if !fastAdd {
			// TODO 不只是此处，还有很多处像这样的特殊情况都需要额外讨论
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

				if err := traceData.CommitRevision(blockVerifyRevision, "block_verify_failed difficulty too low than previous", time.Now()); err != nil {
					bittrace.Error("%v", err)
				}

				return false, false, ruleError(ErrDifficultyTooLow, str)
			}
		}
	}
	if err := traceData.CommitRevision(blockVerifyRevision, "block_verify_success", time.Now()); err != nil {
		bittrace.Error("%v", err)
	}
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
