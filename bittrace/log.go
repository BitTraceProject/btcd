package bittrace

import (
	"github.com/BitTraceProject/BitTrace-Exporter/common"
	"github.com/BitTraceProject/BitTrace-Types/pkg/env"
	"path/filepath"
)

// 初始化 logger

var (
	logger   common.Logger
	envPairs = map[string]string{
		"CONTAINER_NAME":    "",
		"BITTRACE_ROOT_DIR": "",
		"BITTRACE_LOG_DIR":  "",
	}
)

// TODO 替换为最新的库，直接指定 log filepath，这里先暂时用着

func init() {
	err := env.LookupEnvPairs(&envPairs)
	if err != nil {
		panic(err)
	}
	loggerName := envPairs["CONTAINER_NAME"]
	basePath := filepath.Join(envPairs["BITTRACE_ROOT_DIR"], envPairs["BITTRACE_LOG_DIR"])
	logger = common.GetLogger(loggerName, basePath)
}

func Info(format string, msg ...interface{}) {
	logger.Info(format, msg)
}

func Warn(format string, msg ...interface{}) {
	logger.Warn(format, msg)
}

func Error(format string, msg ...interface{}) {
	logger.Error(format, msg)
}

func Fatal(format string, msg ...interface{}) {
	logger.Fatal(format, msg)
}
