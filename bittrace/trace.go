package bittrace

import (
	"sync"
	"time"

	common "github.com/BitTraceProject/BitTrace-Types/pkg/common"
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

func InitSnapshot(targetChainID string, targetChainHeight int32, initTime time.Time) *structure.Snapshot {
	smMux.Lock()
	defer smMux.Unlock()
	snapshot := structure.NewSnapshot(targetChainID, targetChainHeight, initTime)
	snapshotMap[snapshot.ID] = snapshot
	return snapshot
}

func FinalSnapshot(snapshotID string, finalTime time.Time) *structure.Snapshot {
	smMux.Lock()
	defer smMux.Unlock()
	// 这里保证 snapshot 一定存在，所以不检查了
	snapshot := snapshotMap[snapshotID].Commit(finalTime)
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
	gobCodec := common.NewCodecGob(nil)
	rawData, err := gobCodec.Encode(snapshot)
	if err != nil {
		return err
	}
	Data(rawData)
	return nil
}

func (data *TraceData) SetFinalSnapshot(snapshot *structure.Snapshot) error {
	data.Snapshot = snapshot
	gobCodec := common.NewCodecGob(nil)
	rawData, err := gobCodec.Encode(snapshot)
	if err != nil {
		return err
	}
	Data(rawData)
	return nil
}

func (data *TraceData) CommitRevision(revision *structure.Revision, commitTime time.Time, revisionData structure.RevisionData) {
	revision.Commit(commitTime, revisionData)
	data.Snapshot.CommitRevision(revision)
}
