package logger

import (
	"sync"

	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var lock = &sync.Mutex{}

// NewInstance returns a new instance of the underlying
func NewInstance() {
	if logger == nil {
		lock.Lock()
		defer lock.Unlock()
		l, _ := zap.NewProduction()
		logger = l.Sugar()
	}
}

// LogInfo logs the input on the logger's info level
func LogInfo(msg string, params ...interface{}) {
	logger.Infof(msg, params)
}

// LogError logs the input on the logger's error level
func LogError(msg string, params ...interface{}) {
	logger.Errorf(msg, params)
}

// LogWarn logs the input on the logger's warn level
func LogWarn(msg string, params ...interface{}) {
	logger.Warnf(msg, params)
}
