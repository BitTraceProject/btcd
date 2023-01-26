package bittrace

import (
	"encoding/base64"
	"encoding/json"
	"github.com/BitTraceProject/BitTrace-Exporter/common"
	"github.com/BitTraceProject/BitTrace-Types/pkg/constants"
	"github.com/BitTraceProject/BitTrace-Types/pkg/env"
	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// 初始化 logger

var (
	// logger 分离
	prodLogger  common.Logger
	debugLogger common.Logger
	envPairs    = map[string]string{
		"CONTAINER_NAME": "",
	}
	targetHeight = int32(0)

	heightRWMux sync.RWMutex
	syncHeight  = int32(0)
)

func init() {
	err := env.LookupEnvPairs(&envPairs)
	if err != nil {
		panic(err)
	}
	targetHeight, err = getNewTargetHeight()
	if err != nil {
		panic(err)
	}

	// FOR TEST
	targetHeight = 1024

	loggerName := envPairs["CONTAINER_NAME"]
	prodLogger = common.GetLogger(loggerName)
	debugLogger = common.GetLogger(loggerName + "_debug")

	heightRWMux.Lock()
	syncHeight = constants.LOG_SYNC_HEIGHT_INTERVAL // first sync height
	heightRWMux.Unlock()

	go heartbeat()
}

func heartbeat() {
	// 启动一个定时任务，每间隔 1 day 输出一次 sync snapshot
	// 即便达到了 target height 也不停止，这是为了防止 logger 与 exporter 脱轨
	ticker := time.NewTicker(time.Hour * 24) // a day
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			targetChainID := structure.GenChainID(0)
			// 所有字段除了 type 和 timestamp 都是空的
			syncSnapshot := structure.NewSyncSnapshot(targetChainID, 0, time.Now(), "")
			data, err := json.Marshal(syncSnapshot)
			if err != nil {
				debugLogger.Error("[heartbeat]json error:%v", err)
			} else {
				dataBase64 := base64.StdEncoding.EncodeToString(data)
				prodLogger.Msg(dataBase64)
			}
		}
	}
}

func getNewTargetHeight() (int32, error) {
	// get target height
	// get https://blockchain.info/q/getblockcount
	resp, err := http.Get("https://blockchain.info/q/getblockcount") // mainchain
	if err != nil {
		return 0, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err // TODO add default height
	}
	height, err := strconv.ParseInt(string(data), 10, 32)
	return int32(height), nil
}

// Data TODO 添加 height 控制逻辑，根据 height 控制日志输出
func Data(bestHeight int32, data []byte) {
	// 只在 btcd 处控制 height，如果没有达到 height，那么间隔一定的 height，
	// 间隔一 day，向日志写入一个同步标志，其他的消息都忽略，直到达到 height
	//（改 snapshot，btcd Data 函数添加逻辑，其他位置传递 best height，另外 btcd 添加参数 height）

	if ok := dataSync(bestHeight); ok {
		// 到达了 target height
		dataBase64 := base64.StdEncoding.EncodeToString(data)
		prodLogger.Msg(dataBase64)
	}
}

func dataSync(bestHeight int32) bool {
	// 只在没到达 target height 前加锁，返回是否到达 target height，
	// 并且如果没有达到 target height，完成到达 sync height 时同步，并且更新相关字段，
	// 如果 sync height大于等于 target height，那么直接返回 true
	if bestHeight >= targetHeight || syncHeight >= targetHeight {
		return true
	}

	heightRWMux.Lock()
	defer heightRWMux.Unlock()
	if bestHeight >= syncHeight {
		// 到达了 syncHeight，同步
		targetChainID := structure.GenChainID(0)
		syncSnapshot := structure.NewSyncSnapshot(targetChainID, bestHeight, time.Now(), "")
		data, err := json.Marshal(syncSnapshot)
		if err != nil {
			debugLogger.Error("[heartbeat]json error:%v", err)
		} else {
			dataBase64 := base64.StdEncoding.EncodeToString(data)
			prodLogger.Msg(dataBase64)
		}

		syncHeight += constants.LOG_SYNC_HEIGHT_INTERVAL
	}
	return false
}

func Info(format string, msg ...interface{}) {
	debugLogger.Info(format, msg)
}

func Warn(format string, msg ...interface{}) {
	debugLogger.Warn(format, msg)
}

func Error(format string, msg ...interface{}) {
	debugLogger.Error(format, msg)
}

func Fatal(format string, msg ...interface{}) {
	debugLogger.Fatal(format, msg)
}
