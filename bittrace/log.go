package bittrace

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/BitTraceProject/BitTrace-Types/pkg/common"
	"github.com/BitTraceProject/BitTrace-Types/pkg/constants"
	"github.com/BitTraceProject/BitTrace-Types/pkg/logger"
	"github.com/BitTraceProject/BitTrace-Types/pkg/structure"
)

// 初始化 logger

var (
	// logger 分离
	prodLogger  logger.Logger
	debugLogger logger.Logger
	envPairs    = map[string]string{
		"CONTAINER_NAME": "",
		"TARGET_HEIGHT":  "",
	}
	targetHeight = int32(0)

	heightRWMux sync.RWMutex
	syncHeight  = int32(0)
)

func init() {
	common.LookupEnvPairs(&envPairs)
	if envPairs["CONTAINER_NAME"] == "" {
		panic("lookup env CONTAINER_NAME failed")
	}
	loggerName := envPairs["CONTAINER_NAME"]
	prodLogger = logger.GetLogger(loggerName)
	debugLogger = logger.GetLogger(loggerName + "_debug")

	var err error
	if envPairs["TARGET_HEIGHT"] == "" {
		targetHeight, err = getNewTargetHeight()
		if err != nil {
			panic(err)
		}
		debugLogger.Info("[getNewTargetHeight]from api, height=%d", targetHeight)
	} else {
		targetHeight64, err := strconv.ParseInt(envPairs["TARGET_HEIGHT"], 10, 32)
		if err != nil {
			panic(err)
		}
		targetHeight = int32(targetHeight64)
		debugLogger.Info("[getNewTargetHeight]from env, height=%d", targetHeight)
	}

	heightRWMux.Lock()
	syncHeight = constants.LOGGER_SYNC_HEIGHT_INTERVAL // first sync height
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
			targetChainID := common.GenChainID(0)
			// 对于通过 day 定时同步的，所有字段除了 type 和 timestamp 都是空的，state 为 nil
			syncSnapshot := structure.NewSyncSnapshot(targetChainID, 0, time.Now(), nil)
			debugLogger.Info("[heartbeat]%+v", syncSnapshot)
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}
	resp, err := c.Get("https://blockchain.info/q/getblockcount") // mainchain
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

func Data(data []byte, bestState *structure.BestState) {
	if ok := dataSync(bestState); ok {
		// 到达了 target height
		dataBase64 := base64.StdEncoding.EncodeToString(data)
		prodLogger.Msg(dataBase64)
	}
}

func dataSync(bestState *structure.BestState) bool {
	// 只在没到达 target height 前加锁，返回是否到达 target height，
	// 并且如果没有达到 target height，完成到达 sync height 时同步，并且更新相关字段，
	// 如果 sync height大于等于 target height，那么直接返回 true
	if bestState.Height >= targetHeight || syncHeight >= targetHeight {
		return true
	}

	heightRWMux.Lock()
	defer heightRWMux.Unlock()
	if bestState.Height >= syncHeight {
		// 到达了 syncHeight，同步
		targetChainID := common.GenChainID(0)
		// 对于通过 height 间隔同步的，所有字段都是正常的，state 也不为 nil
		syncSnapshot := structure.NewSyncSnapshot(targetChainID, bestState.Height, time.Now(), bestState)
		debugLogger.Info("[dataSync]%+v", syncSnapshot)
		data, err := json.Marshal(syncSnapshot)
		if err != nil {
			debugLogger.Error("[dataSync]json error:%v", err)
		} else {
			dataBase64 := base64.StdEncoding.EncodeToString(data)
			prodLogger.Msg(dataBase64)
		}
		syncHeight += constants.LOGGER_SYNC_HEIGHT_INTERVAL
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
