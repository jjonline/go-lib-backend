package crond

import (
	"github.com/robfig/cron/v3"
	"log/slog"
	"sync"
	"time"
)

// 定义日志字段中标记类型的名称
const module = "crontab"

// Crontab 定时任务实现
type Crontab struct {
	cron   *cron.Cron   // 定时任务实例
	logger *slog.Logger // 日志输出
	lock   sync.Mutex   // 并发锁
}

// registeredCommand 已注册的定时任务映射map
var registeredCommand = make(map[int]CronTask)

// New 实例化crontab实例
func New(logger *slog.Logger) *Crontab {
	log := cronLog{logger: *logger}
	timeZone := time.FixedZone("Asia/Shanghai", 8*3600) // 东八区
	return &Crontab{
		cron:   cron.New(cron.WithSeconds(), cron.WithLogger(log), cron.WithLocation(timeZone)),
		logger: logger,
		lock:   sync.Mutex{},
	}
}

// Register 注册定时任务类
func (c *Crontab) Register(task CronTask) {
	c.lock.Lock()
	defer c.lock.Unlock()

	loggerWithAttr := c.logger.With(
		slog.String("module", module),
		slog.String("signature", task.Signature()),
		slog.String("description", task.Desc()),
		slog.String("rule", task.Rule()),
	)

	// 任务类包装
	wrapper := func() {
		// 处理并恢复业务代码可能导致的panic，避免cron进程退出
		defer func() {
			if err := recover(); err != nil {
				// record panic log
				loggerWithAttr.Error(
					"crontab.panic",
					slog.Any("error", err),
				)
			}
		}()

		// 执行定时任务
		loggerWithAttr.Info("crontab.execute.start")
		err := task.Execute()
		if err != nil {
			loggerWithAttr.Error("crontab.execute.failed", slog.Any("error", err))
		} else {
			loggerWithAttr.Info("crontab.execute.ok")
		}
	}

	// 注册任务
	entryId, err := c.cron.AddFunc(task.Rule(), wrapper)
	if err != nil {
		loggerWithAttr.Error("crontab.register.err", slog.Any("error", err))
	} else {
		loggerWithAttr.Info("crontab.register.ok")
		registeredCommand[int(entryId)] = task
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
