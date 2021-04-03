package crontab

import (
	"github.com/jjonline/go-mod-library/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// cronLog cron日志记录器
type cronLog struct {
	logger *logger.Logger
}

// Info
func (l cronLog) Info(msg string, keysAndValues ...interface{}) {
	// record
	fields := []zapcore.Field{}
	for key, val := range keysAndValues {
		if "now" == val {
			fields = append(fields, zap.Any("execute_time", keysAndValues[key+1]))
		}
		if "entry" == val {
			entryID := keysAndValues[key+1]
			fields = append(fields, zap.Any("entry_id", entryID))
		}
		if "next" == val {
			fields = append(fields, zap.Any("next_time", keysAndValues[key+1]))
		}
	}

	switch msg {
	case "wake":
	case "schedule":
		l.logger.Zap.Debug("crontab", fields...)
	default:
		l.logger.Zap.Info("crontab", fields...)
	}
}

// Error
func (l cronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	fields := []zapcore.Field{}
	fields = append(fields, zap.Error(err), zap.String("msg", msg))
	for key, val := range keysAndValues {
		if "stack" == val {
			fields = append(fields, zap.Any("stack", keysAndValues[key+1]))
		}
	}
	l.logger.Zap.Error("crontab", fields...)
}
