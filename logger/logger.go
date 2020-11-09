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

// Info logs the input on the logger's info level
func Info(msg string, params ...interface{}) {
	if logger == nil {
		return
	}

	if len(params) <= 0 {
		logger.Info(msg)
	}

	logger.Infof(msg, params)
}

// Error logs the input on the logger's error level
func Error(msg string, params ...interface{}) {
	if logger == nil {
		return
	}

	if len(params) <= 0 {
		logger.Error(msg)
	}

	logger.Errorf(msg, params)
}

// Warn logs the input on the logger's warn level
func Warn(msg string, params ...interface{}) {
	if logger == nil {
		return
	}

	if len(params) <= 0 {
		logger.Warn(msg)
	}

	logger.Warnf(msg, params)
}
