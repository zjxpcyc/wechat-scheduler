package lib

import (
	"github.com/zjxpcyc/tinylogger"
)

var logger tinylogger.LogService

// GetLogger 获取日志记录器
func GetLogger() tinylogger.LogService {
	return logger
}

func init() {
	logger = new(tinylogger.Logger)
}
