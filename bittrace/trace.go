package bittrace

import (
	"encoding/json"
	"github.com/btcsuite/btcd/blockchain"
	"sync"
	"time"

	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
)

type TraceData struct {
	*structure.Snapshot
}

var (
	smMux       sync.Mutex
	snapshotMap map[string]*structure.Snapshot
)

func init() {
	snapshotMap = map[string]*structure.Snapshot{}
}

func InitSnapshot(targetChainID string, targetChainHeight int32, initTime time.Time, bestState *blockchain.BestState) *structure.Snapshot {
	smMux.Lock()
	defer smMux.Unlock()
	state := &structure.BestState{
		Hash:            bestState.Hash.String(),
		Height:          bestState.Height,
		Bits:            bestState.Bits,
		BlockSize:       bestState.BlockSize,
		BlockWeight:     bestState.BlockWeight,
		NumTxns:         bestState.NumTxns,
		TotalTxns:       bestState.TotalTxns,
		MedianTimestamp: structure.FromTime(bestState.MedianTime),
	}
	snapshot := structure.NewInitSnapshot(targetChainID, targetChainHeight, initTime, state)
	snapshotMap[snapshot.ID] = snapshot
	return snapshot
}

func FinalSnapshot(snapshotID string, finalTime time.Time, bestState *blockchain.BestState) *structure.Snapshot {
	smMux.Lock()
	defer smMux.Unlock()
	// 这里保证 snapshot 一定存在，所以不检查了
	state := &structure.BestState{
		Hash:            bestState.Hash.String(),
		Height:          bestState.Height,
		Bits:            bestState.Bits,
		BlockSize:       bestState.BlockSize,
		BlockWeight:     bestState.BlockWeight,
		NumTxns:         bestState.NumTxns,
		TotalTxns:       bestState.TotalTxns,
		MedianTimestamp: structure.FromTime(bestState.MedianTime),
	}
	snapshot := snapshotMap[snapshotID].Commit(finalTime, state)
	// 删掉 snapshot 临时副本，释放内存
	delete(snapshotMap, snapshotID)
	return snapshot
}

func NewTraceData() *TraceData {
	return &TraceData{
		Snapshot: nil,
	}
}

func (data *TraceData) SetInitSnapshot(snapshot *structure.Snapshot) error {
	data.Snapshot = snapshot
	rawData, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	bestState := snapshot.State
	Data(rawData, bestState)
	return nil
}

func (data *TraceData) SetFinalSnapshot(snapshot *structure.Snapshot) error {
	data.Snapshot = snapshot
	rawData, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	bestState := snapshot.State
	Data(rawData, bestState)
	return nil
}

func (data *TraceData) CommitRevision(revision *structure.Revision, commitTime time.Time, revisionData structure.RevisionData) {
	revision.Commit(commitTime, revisionData)
	data.Snapshot.CommitRevision(revision)
}
