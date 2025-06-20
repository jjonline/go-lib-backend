package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Options 参数选项
type Options struct {
	// Target 日志存储目标，支持文件目录路径：./runtime/、/opt/logs/ 或 实现 io.Writer 的写入器例如: os.Stdout 、 os.Stderr
	Target any
	// AddSource 日志是否添加代码source，默认false表示不添加
	AddSource bool
	// UseText 日志格式使用text文本，默认false表示使用json
	UseText bool
	// MaxSize 当日志 Target 为文件目录时支持日志轮转切割，指定单个文件最大体积，超过自动切割，设置0表示不按文件体积轮转，单位：比特
	MaxSize int64
	// MaxDays 当日志 Target 为文件时支持日志轮转和切割，指定保留日志文件的最大天数，设置0表示不清理
	MaxDays int64
}

// New 初始化单例logger
//
//	-- opt 参考 Options 结构体说明，可给nil默认输出到标准输出
//
//		logger := logger.New(&logger.Options{Target: os.Stdout})
//		// can render info log
//		logger.Info("info", "info", "testing")
//		// change log level
//		logger.GetSlogLeveler().Set(slog.LevelInfo)
//		// none log render
//		logger.Info("info", "info", "testing")
//		// high performance
//		logger.Info("info", slog.String("info", "testing"))
//		// get slog.Logger instance
//		slogLogger := logger.GetSlogLogger()
//		// get log.Logger instance
//		logger.GetLogLogger()
//	使用原生log/slog即可，无需引入第三方包
//	默认日志级别为slog.LevelWarn，可通过logger.GetSlogLeveler().Set(slog.LevelInfo)重设
func New(opt *Options) *Logger {
	if nil == opt {
		opt = &Options{
			Target:    os.Stdout,
			AddSource: false,
			UseText:   false,
			MaxSize:   0,
			MaxDays:   0,
		}
	}

	// deal options
	var (
		level = &slog.LevelVar{}
	)

	// deal target writer
	var writer io.Writer
	switch t := opt.Target.(type) {
	case io.Writer:
		writer = t
	case string:
		switch t {
		case "stdout":
			writer = os.Stdout
		case "stderr":
			writer = os.Stderr
		default:
			if !strings.HasSuffix(t, "/") {
				t += "/"
			}
			dir := filepath.Dir(t)
			if !checkFileExist(dir) {
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					panic(fmt.Errorf("%w", err))
				}
			}
			// rotate writer
			writer = newDailySizeRotateWriter(dir, opt.MaxSize, opt.MaxDays)
		}
	default:
		panic("unsupported target type, support string and io.Writer, string can use stdout as io.Stdout or stderr as os.Stderr or DIR path with suffix slash")
	}

	// default level Warning
	level.Set(slog.LevelWarn)

	// deal slog Handler
	var handler slog.Handler
	if opt.UseText {
		// use text handler
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			AddSource:   opt.AddSource,
			Level:       level,
			ReplaceAttr: nil,
		})
	} else {
		// use json handler
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			AddSource:   opt.AddSource,
			Level:       level,
			ReplaceAttr: nil,
		})
	}

	logger := &Logger{
		writer:  writer,
		level:   level,
		handler: handler,
		logger:  slog.New(handler),
	}

	// set slog default logger use this instance
	slog.SetDefault(logger.logger)

	return logger
}

// Logger logger封装
type Logger struct {
	writer  io.Writer
	handler slog.Handler
	level   *slog.LevelVar
	logger  *slog.Logger
}

// Debug debug level log
func (l *Logger) Debug(msg string, keyValue ...any) {
	l.logger.Debug(msg, keyValue...)
}

// Info info level log
func (l *Logger) Info(msg string, keyValue ...any) {
	l.logger.Info(msg, keyValue...)
}

// Warn warn level log
func (l *Logger) Warn(msg string, keyValue ...any) {
	l.logger.Warn(msg, keyValue...)
}

// Error error level log
func (l *Logger) Error(msg string, keyValue ...any) {
	l.logger.Error(msg, keyValue...)
}

// GetSlogLogger 获取 slog.Logger 实例
func (l *Logger) GetSlogLogger() *slog.Logger {
	return l.logger
}

// GetLogLogger 获取 log.Logger 实例
func (l *Logger) GetLogLogger() *log.Logger {
	return slog.NewLogLogger(l.handler, l.level.Level())
}

// GetWriter 获取底层writer实现
func (l *Logger) GetWriter() io.Writer {
	return l.writer
}

// GetSlogLeveler 获取底层slog.Leveler，可自定义日志级别
//
//	默认日志级别为slog.LevelWarn， 示例：logger.GetSlogLeveler().Set(slog.LevelInfo)
func (l *Logger) GetSlogLeveler() *slog.LevelVar {
	return l.level
}
