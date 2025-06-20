package crond

import (
	"github.com/robfig/cron/v3"
	"log/slog"
)

// cronLog cron日志记录器
type cronLog struct {
	logger slog.Logger
}

// Info 信息级别日志输出
func (l cronLog) Info(msg string, keysAndValues ...interface{}) {
	// record
	var fields []any
	fields = append(fields, slog.String("module", module))
	for key, val := range keysAndValues {
		if "now" == val {
			fields = append(fields, slog.Any("execute_time", keysAndValues[key+1]))
		}
		if "entry" == val {
			entryID := keysAndValues[key+1]
			entryIntID := int(entryID.(cron.EntryID))
			fields = append(fields, slog.Any("entry_id", entryIntID))

			// 取当前command的签名添加日志字段
			if command, exist := registeredCommand[entryIntID]; exist {
				fields = append(fields, slog.Any("signature", command.Signature()))
			}
		}
		if "next" == val {
			fields = append(fields, slog.Any("next_time", keysAndValues[key+1]))
		}
	}

	// 忽略掉wake类型
	if msg != "wake" {
		l.logger.Info("crontab.log."+msg, fields...)
	}
}

// Error 错误级别日志输出
func (l cronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	var fields []any
	fields = append(fields, slog.String("module", module))
	if err != nil {
		slog.String("error", err.Error())
	}
	for key, val := range keysAndValues {
		if "stack" == val {
			fields = append(fields, slog.Any("stack", keysAndValues[key+1]))
		}
	}
	l.logger.Error("crontab.error."+msg, fields...)
}
