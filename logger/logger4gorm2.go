package logger

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

// gorm2Logger gorm2的日志自定义实现
type gorm2Logger struct {
	level    zapcore.Level // 转换为zap的日志级别予以判断
	logger   zap.Logger
	levelMap map[logger.LogLevel]zapcore.Level
}

// LogMode 设置gorm的日志级别，内部转换为zap的日志级别用于判断
func (g gorm2Logger) LogMode(level logger.LogLevel) logger.Interface {
	g.level = g.levelMap[level]
	return &g
}

// Info 输出info级别的日志
// gorm里只有一些添加callback处理函数的钩子时触发<譬如update时自定义钩子>，内部调整为启用debug级别才记录日志
func (g gorm2Logger) Info(ctx context.Context, s string, i ...interface{}) {
	if g.logger.Core().Enabled(zap.DebugLevel) {
		g.logger.Info(
			"gorm.log",
			zap.String("module", "gorm.info"),
			zap.String("info", fmt.Sprintf(s, i[0])),
			zap.String("source", i[1].(string)), // 输出日志的文件和行数
		)
	}
}

// Warn 输出警告日志
// 1、gorm里移除、重排callback处理函数的钩子时触发<譬如update时自定义钩子>，内部调整为启用debug级别才记录日志
// 2、自定义model不匹配时，内部调整为启用debug级别才记录日志
func (g gorm2Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	if g.logger.Core().Enabled(zap.DebugLevel) {
		switch len(i) {
		case 2:
			g.logger.Info(
				"gorm.log",
				zap.String("module", "gorm.info"),
				zap.String("info", fmt.Sprintf(s, i[0])),
				zap.String("source", i[1].(string)), // 输出日志的文件和行数
			)
		case 3:
			g.logger.Info(
				"gorm.log",
				zap.String("module", "gorm.info"),
				zap.String("info", fmt.Sprintf(s, i[0], i[1])),
				zap.String("source", i[2].(string)),
			)
		}
	}
}

// Error 输出出错日志
// gorm内部一些执行的错误：初始化db、解析schema、解析自定义回调钩子、dryRun出错时，内部调整为启用info级别才记录日志
func (g gorm2Logger) Error(ctx context.Context, s string, i ...interface{}) {
	if g.logger.Core().Enabled(zap.InfoLevel) {
		g.logger.Info(
			"gorm.log",
			zap.String("module", "gorm.info"),
			zap.String("info", fmt.Sprintf(s, i...)),
		)
	}
}

// Trace 输出trace级别日志
// gorm内部主要用于输出执行的sql日志，调整为启用debug级别才输出日志
func (g gorm2Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.logger.Core().Enabled(zap.DebugLevel) {
		elapsed := time.Since(begin) // 执行耗时
		sql, rows := fc()            // 执行的sql和影响行数

		if err != nil {
			g.logger.Error(
				"gorm.log",
				zap.String("module", "gorm.sql"),
				zap.Error(err), // 注意：执行查询未找到记录也会在此
				zap.String("sql", sql),
				zap.Int64("affect_rows", rows),
				zap.String("source", utils.FileWithLineNum()),
				zap.Duration("duration", elapsed),
			)
		} else {
			g.logger.Error(
				"gorm.log",
				zap.String("module", "gorm.sql"),
				zap.String("sql", sql),
				zap.Int64("affect_rows", rows),
				zap.String("source", utils.FileWithLineNum()),
				zap.Duration("duration", elapsed),
			)
		}
	}
}

// NewGorm2Logger 创建gorm2的logger实例
func NewGorm2Logger() logger.Interface {
	if zapLogger == nil {
		panic(errors.New("please use logger.New() create logger.Logger instance at first"))
	}

	return gorm2Logger{logger: *zapLogger}
}
