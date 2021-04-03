package crontab

import (
	"github.com/jjonline/go-mod-library/contract"
	"github.com/jjonline/go-mod-library/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Crontab
type Crontab struct {
	commands map[int]contract.Command // 注册的所有定时任务
	cron     *cron.Cron               // 定时任务实例
	logger   *logger.Logger           // 日志输出
	lock     sync.Mutex               // 并发锁
}

// New 实例化crontab实例
func New(logger *logger.Logger) *Crontab {
	log := cronLog{logger: logger}
	timeZone := time.FixedZone("Asia/Shanghai", 8*3600) // 东八区
	return &Crontab{
		commands: make(map[int]contract.Command),
		cron:     cron.New(cron.WithSeconds(), cron.WithLogger(log), cron.WithLocation(timeZone)),
		logger:   logger,
		lock:     sync.Mutex{},
	}
}

// Register 注册定时任务类
// @param spec string 定时规则：`Second | Minute | Hour | Dom (day of month) | Month | Dow (day of week)`
// @param command contract.Command 任务类需实现命令契约，并且传递结构体实例的指针
func (c *Crontab) Register(spec string, command contract.Command) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 任务类包装
	wrapper := func() {
		// 处理并恢复业务代码可能导致的panic，避免cron进程退出
		defer func() {
			if err := recover(); err != nil {
				// record log
				c.logger.Zap.Error(
					"crontab.panic",
					zap.String("signature", command.Signature()),
					zap.String("description", command.Description()),
					zap.Stack("stack"),
				)
			}
		}()

		// 执行定时任务
		if err := command.Execute(); err != nil {
			c.logger.Zap.Error(
				"crontab.error",
				zap.String("signature", command.Signature()),
				zap.String("description", command.Description()),
				zap.Stack("stack"),
			)
		} else {
			c.logger.Zap.Info(
				"crontab.execute",
				zap.String("signature", command.Signature()),
				zap.String("description", command.Description()),
				zap.String("spec", spec),
			)
		}
	}

	// 注册任务
	entry_id, err := c.cron.AddFunc(spec, wrapper)
	if err == nil {
		c.logger.Zap.Info(
			"crontab.register",
			zap.String("signature", command.Signature()),
			zap.String("description", command.Description()),
			zap.String("spec", spec),
		)
		c.commands[int(entry_id)] = command
	}
}

// Start 启动定时任务守护进程
func (c *Crontab) Start() {
	c.logger.Info("crontab.started")
	c.cron.Start()
}

// Shutdown 优雅停止定时任务守护进程
func (c *Crontab) Shutdown() error {
	c.cron.Stop()
	c.logger.Info("crontab.stopped")
	return nil
}
