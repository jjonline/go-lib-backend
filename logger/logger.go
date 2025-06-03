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
	AddSource bool  // 日志是否添加代码source，默认false表示不添加
	UseText   bool  // 日志格式使用text文本，默认false表示使用json
	MaxSize   int64 // 当日志target为文件时支持日志轮转和切割，此处指定单个文件最大体积，超过最大体积会自动切割，单位：比特，不设置或设置0表示不按文件体积切割
	MaxDays   int64 // 当日志target为文件时支持日志轮转和切割，指定保留日志文件的最大天数，不设置或设置0表示不清理
}

// New 初始化单例logger
//
//	-- level   日志级别：debug、info、warning 等，传入变量可实时控制调整日志级别
//	-- target  文件路径形式的目录路径：./runtime/、/opt/logs/ 或 字符串stderr、stdout表示标准输出 或 实现 io.Writer 的写入器
//	-- useText 是否使用文本格式，默认false，默认json格式日志
//
//		lvl := &slog.LevelVar{}
//		lvl.Set(slog.LevelInfo)
//		logger := logger.New(lvl, "stdout")
//		// logger := logger.New(lvl, os.Stdout)
//		// can render info log
//		logger.Info("info", "info", "testing")
//		// change log level
//		lvl.Set(slog.LevelWarn)
//		// none log render
//		logger.Info("info", "info", "testing")
//		// high performance
//		logger.Info("info", slog.String("info", "testing"))
//		// get slog.Logger instance
//		slogLogger := logger.GetSlogLogger()
//		// get log.Logger instance
//		logger.GetLogLogger()
//	使用原生log/slog即可，无需引入第三方包
func New(level *slog.LevelVar, target any, opts ...Options) *Logger {
	// deal options
	var opt = Options{
		UseText: false,
		MaxSize: 0,
		MaxDays: 0,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// deal target writer
	var writer io.Writer
	switch t := target.(type) {
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
			// 默认512M自动rotate
			writer = newDailySizeRotateWriter(dir, opt.MaxSize, opt.MaxDays)
		}
	default:
		panic("unsupported target type, support string and io.Writer, string can use stdout as io.Stdout or stderr as os.Stderr or DIR path with suffix slash")
	}

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
