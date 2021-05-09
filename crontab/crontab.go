package crontab

import (
	"github.com/jjonline/go-mod-library/contract"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Crontab
type Crontab struct {
	cron   *cron.Cron  // 定时任务实例
	logger *zap.Logger // 日志输出
	lock   sync.Mutex  // 并发锁
}

//  registeredCommand 已注册的定时任务映射map
var registeredCommand = make(map[int]contract.Command)

// New 实例化crontab实例
func New(logger *zap.Logger) *Crontab {
	log := cronLog{logger: logger}
	timeZone := time.FixedZone("Asia/Shanghai", 8*3600) // 东八区
	return &Crontab{
		cron:   cron.New(cron.WithSeconds(), cron.WithLogger(log), cron.WithLocation(timeZone)),
		logger: logger,
		lock:   sync.Mutex{},
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
				c.logger.Error(
					"crontab.panic",
					zap.String("signature", command.Signature()),
					zap.String("description", command.Description()),
					zap.Stack("stack"),
				)
			}
		}()

		// 执行定时任务
		if err := command.Execute(); err != nil {
			c.logger.Error(
				"crontab.execute.failed",
				zap.String("signature", command.Signature()),
				zap.String("description", command.Description()),
				zap.String("spec", spec),
			)
		}
	}

	// 注册任务
	entry_id, err := c.cron.AddFunc(spec, wrapper)
	if err != nil {
		c.logger.Error(
			"crontab.register.err",
			zap.String("signature", command.Signature()),
			zap.String("description", command.Description()),
			zap.String("spec", spec),
			zap.Error(err),
		)
	} else {
		c.logger.Info(
			"crontab.register",
			zap.String("signature", command.Signature()),
			zap.String("description", command.Description()),
			zap.String("spec", spec),
		)
		registeredCommand[int(entry_id)] = command
	}
}

// Start 启动定时任务守护进程
func (c *Crontab) Start() {
	c.cron.Start()
}

// Shutdown 优雅停止定时任务守护进程
func (c *Crontab) Shutdown() {
	c.cron.Stop()
}
