package bittrace

import "log"

//import (
//	bitlog_common "github.com/1uvu/bitlog/pkg/common"
//)
//
//// 初始化 logger
//
//var (
//	logger    bitlog_common.Logger
//	constants = map[bitlog_common.ConstantKey]string{
//		bitlog_common.ROOT_DIR:   "", // TODO 从 env 读取下面的值
//		bitlog_common.CLIENT_DIR: "",
//		bitlog_common.LOG_DIR:    "",
//		bitlog_common.CONFIG_DIR: "",
//	}
//)
//
//// TODO 替换为最新的库，直接指定 log filepath，这里先暂时用着
//
//func init() {
//	bitlog_common.InitConstants(constants)
//	logger = bitlog_common.GetLogger("btcd_logger", bitlog_common.GetConstants(bitlog_common.LOG_DIR))
//}
//
//func Info(format string, msg ...interface{}) {
//	logger.Info(format, msg)
//}
//
//func Warn(format string, msg ...interface{}) {
//	logger.Warn(format, msg)
//}
//
//func Error(format string, msg ...interface{}) {
//	logger.Error(format, msg)
//}
//
//func Fatal(format string, msg ...interface{}) {
//	logger.Fatal(format, msg)
//}

func Info(format string, msg ...interface{}) {
	log.Printf(format, msg)
}

func Warn(format string, msg ...interface{}) {
	log.Printf(format, msg)
}

func Error(format string, msg ...interface{}) {
	log.Printf(format, msg)
}

func Fatal(format string, msg ...interface{}) {
	log.Printf(format, msg)
}
