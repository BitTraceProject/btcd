package bittrace

import (
	"encoding/base64"
	"github.com/BitTraceProject/BitTrace-Exporter/common"
	"github.com/BitTraceProject/BitTrace-Types/pkg/env"
)

// 初始化 logger

var (
	// logger 分离
	prodLogger  common.Logger
	debugLogger common.Logger
	envPairs    = map[string]string{
		"CONTAINER_NAME": "",
	}
)

func init() {
	err := env.LookupEnvPairs(&envPairs)
	if err != nil {
		panic(err)
	}
	loggerName := envPairs["CONTAINER_NAME"]
	prodLogger = common.GetLogger(loggerName)
	debugLogger = common.GetLogger(loggerName + "_debug")
}

func Data(data []byte) {
	logData := base64.StdEncoding.EncodeToString(data)
	prodLogger.Msg(logData)
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
