package logger

import (
	"fmt"
	"go.uber.org/zap"
)

// Logger logger封装, 实现第三方库的日志接口
type Logger struct {
	Zap *zap.Logger
}

// New 初始化单例logger
// @param level 日志级别：debug、info、warning 等
// @param path  文件形式的日志路径 or 标准输出 stderr
func New(level, path string) *Logger {
	return &Logger{
		Zap: newZap(level, path),
	}
}

func (l Logger) Debug(msg string) {
	zapLogger.Debug(msg)
}
func (l Logger) Info(msg string) {
	zapLogger.Info(msg)
}
func (l Logger) Warn(msg string) {
	zapLogger.Warn(msg)
}
func (l Logger) Error(msg string) {
	zapLogger.Error(msg)
}
func (l Logger) Debugf(format string, args ...interface{}) {
	zapLogger.Debug(fmt.Sprintf(format, args...))
}
func (l Logger) Infof(format string, args ...interface{}) {
	zapLogger.Info(fmt.Sprintf(format, args...))
}
func (l Logger) Warnf(format string, args ...interface{}) {
	zapLogger.Warn(fmt.Sprintf(format, args...))
}
func (l Logger) Errorf(format string, args ...interface{}) {
	zapLogger.Error(fmt.Sprintf(format, args...))
}
func (l Logger) Print(v ...interface{}) {
	zapLogger.Info(fmt.Sprint(v...))
}
func (l Logger) Printf(format string, v ...interface{}) {
	zapLogger.Info(fmt.Sprintf(format, v...))
}
