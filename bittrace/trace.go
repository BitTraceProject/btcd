package bittrace

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
)

type TraceData struct {
	*structure.Snapshot
}

var (
	smMux       sync.Mutex
	snapshotMap map[string]structure.Snapshot

	fsMux       sync.RWMutex
	finalStatus structure.Status
)

// TODO 这里初始化需要弄，初始化状态，初始化链 id 等等，在 client 启动的某个位置添加
func init() {
	snapshotMap = map[string]structure.Snapshot{}
}

func InitSnapshot(targetChainID string, targetChainHeight int32, initTime time.Time, initStatus structure.Status) structure.Snapshot {
	smMux.Lock()
	defer smMux.Unlock()
	snapshot := structure.InitSnapshot(targetChainID, targetChainHeight, initTime, initStatus)
	snapshotMap[snapshot.ID] = snapshot
	return snapshot
}

func FinalSnapshot(snapshotID string, finalTime time.Time, finalStatus structure.Status) structure.Snapshot {
	smMux.Lock()
	defer smMux.Unlock()
	// 这里保证 snapshot 一定存在
	snapshot := structure.FinalSnapshot(snapshotMap[snapshotID], finalTime, finalStatus)
	// 删掉 snapshot 临时副本，释放内存
	delete(snapshotMap, snapshotID)
	return snapshot
}

func InitFinalStatus(status structure.Status) {
	finalStatus = status
}

func GetFinalStatus() structure.Status {
	fsMux.RLock()
	defer fsMux.RUnlock()
	return finalStatus
}

func UpdateFinalStatus(status structure.Status) {
	fsMux.Lock()
	defer fsMux.Unlock()
	finalStatus = status
}

func NewTraceData() *TraceData {
	return &TraceData{
		Snapshot: nil,
	}
}

func (data *TraceData) SetInitSnapshot(snapshot *structure.Snapshot) error {
	data.Snapshot = snapshot
	rawData, err := json.Marshal(*snapshot)
	if err != nil {
		return err
	}
	Data(rawData)
	return nil
}

func (data *TraceData) SetFinalSnapshot(snapshot *structure.Snapshot) error {
	data.Snapshot = snapshot
	rawData, err := json.Marshal(*snapshot)
	if err != nil {
		return err
	}
	Data(rawData)
	return nil
}

func (data *TraceData) CommitRevision(revision *structure.Revision, context string, commitTime time.Time) error {
	err := revision.Commit(context, commitTime)
	if err != nil {
		return err
	}
	data.Snapshot.RevisionList = append(data.Snapshot.RevisionList, *revision)
	return nil
}

func (data *TraceData) LastRevision() *structure.Revision {
	return &data.Snapshot.RevisionList[len(data.Snapshot.RevisionList)-1]
}
