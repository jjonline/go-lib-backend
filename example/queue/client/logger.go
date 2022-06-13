package client

import (
	"github.com/jjonline/go-lib-backend/logger"
	"go.uber.org/zap"
)

// logger
var zapLogger = logger.New("debug", "stdout", "module")

// DefineLogger 实现queue要求的logger
type DefineLogger struct {
}

func (d *DefineLogger) toZapFiled(keyValue []string) []zap.Field {
	if len(keyValue) > 0 {
		var res = make([]zap.Field, 0)
		for i := 0; i < len(keyValue); i = i + 2 {
			res = append(res, zap.String(keyValue[i], keyValue[i+1]))
		}
		return res
	}
	return nil
}

func (d *DefineLogger) Debug(msg string, keyValue ...string) {
	zapLogger.Zap.Debug(msg, d.toZapFiled(keyValue)...)
}

func (d *DefineLogger) Info(msg string, keyValue ...string) {
	zapLogger.Zap.Info(msg, d.toZapFiled(keyValue)...)
}

func (d *DefineLogger) Warn(msg string, keyValue ...string) {
	zapLogger.Zap.Warn(msg, d.toZapFiled(keyValue)...)
}

func (d *DefineLogger) Error(msg string, keyValue ...string) {
	zapLogger.Zap.Error(msg, d.toZapFiled(keyValue)...)
}
