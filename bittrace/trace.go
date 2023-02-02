package bittrace

import (
	"encoding/json"
	"time"

	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
)

type TraceData struct {
	initSnapshot  *structure.Snapshot
	finalSnapshot *structure.Snapshot
}

func NewTraceData() *TraceData {
	return &TraceData{
		initSnapshot:  nil,
		finalSnapshot: nil,
	}
}

func (data *TraceData) SetInitSnapshot(targetChainID string, targetChainHeight int32, initTime time.Time, bestState *structure.BestState) error {
	snapshot := structure.NewInitSnapshot(targetChainID, targetChainHeight, initTime, bestState)
	data.initSnapshot = snapshot
	rawData, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	Data(rawData, bestState)
	return nil
}

func (data *TraceData) SetFinalSnapshot(finalTime time.Time, bestState *structure.BestState) error {
	snapshot := data.initSnapshot.Commit(finalTime, bestState)
	data.finalSnapshot = snapshot
	rawData, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	Data(rawData, bestState)
	return nil
}

func (data *TraceData) CommitRevision(revision *structure.Revision, commitTime time.Time, revisionData structure.RevisionData) {
	revision.Commit(commitTime, revisionData)
	data.initSnapshot.CommitRevision(revision)
}

func (data *TraceData) GetInitSnapshot() *structure.Snapshot {
	return data.initSnapshot
}

func (data *TraceData) GetFinalSnapshot() *structure.Snapshot {
	return data.finalSnapshot
}

func (data *TraceData) GetSnapshotID() string {
	return data.initSnapshot.ID
}
