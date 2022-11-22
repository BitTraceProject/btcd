package bittrace

import (
	"sync"
	"time"

	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
)

type TraceData struct {
	initSnapshot  *structure.Snapshot
	finalSnapshot *structure.Snapshot
	revisionList  []*structure.Revision
}

var (
	smMux       sync.Mutex
	snapshotMap map[string]structure.Snapshot

	fsMux       sync.RWMutex
	finalStatus structure.Status
)

// TODO 这里初始化需要弄，初始化状态，初始化链 id 等等
func init() {
	snapshotMap = map[string]structure.Snapshot{}
	finalStatus = structure.NewStatus(nil, nil)
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
		initSnapshot: nil,
		revisionList: []*structure.Revision{},
	}
}

func (data *TraceData) SetInitSnapshot(snapshot *structure.Snapshot) {
	data.initSnapshot = snapshot
	Info("got a init snapshot:[%+v]", *snapshot)
}

func (data *TraceData) CurrentInitSnapshot() *structure.Snapshot {
	return data.initSnapshot
}

func (data *TraceData) SetFinalSnapshot(snapshot *structure.Snapshot) {
	data.finalSnapshot = snapshot
	Info("got a final snapshot:[%+v]", *snapshot)
}

func (data *TraceData) CurrentFinalSnapshot() *structure.Snapshot {
	return data.finalSnapshot
}

func (data *TraceData) AddRevision(revision *structure.Revision) {
	data.revisionList = append(data.revisionList, revision)
	Info("add a revision:[%+v]", *revision)
	// TODO 这里评估下是否可以直接加到 snapshot，一次性输出完毕
}

func (data *TraceData) LastRevision() *structure.Revision {
	return data.revisionList[len(data.revisionList)-1]
}
