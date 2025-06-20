package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jjonline/go-lib-backend/example/queue/client"
	"github.com/jjonline/go-lib-backend/example/queue/tasks"
	"github.com/jjonline/go-lib-backend/logger"
	"github.com/jjonline/go-lib-backend/queue"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	_logger := logger.New(nil)
	_logger.GetSlogLeveler().Set(slog.LevelInfo)

	// 使用memory内存驱动
	// !!!警告：本地memory驱动仅能用于本地开发调试，不得用于prod生产环境，此处仅为示例!!!
	_logger.Info("init queue service")
	queueService := queue.New(queue.Redis, client.NewRedis(), _logger, 5)

	// 或者使用 redis驱动：请留意redis的链接信息，示例使用了本机redis
	// redisClient := client.NewRedis()
	// queueService := queue.New(queue.Memory, nil, _logger, 5)

	// register task
	_logger.Info("register task")
	//_ = queueService.BootstrapOne(&tasks.TestTask{})
	_ = queueService.BootstrapOne(&tasks.TestTimeout{})

	idleCloser := make(chan struct{})

	// 接收退出信号
	quitChan := make(chan os.Signal)
	signal.Notify(
		quitChan,
		syscall.SIGINT,  // 用户发送INTR字符(Ctrl+C)触发
		syscall.SIGTERM, // 结束程序
		syscall.SIGHUP,  // 终端控制进程结束(终端连接断开)
		syscall.SIGQUIT, // 用户发送QUIT字符(Ctrl+/)触发
	)

	go func() {
		// wait exit signal
		<-quitChan

		_logger.Info("receive exit signal")

		// shutdown worker daemon with timeout context
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// graceful shutdown by signal
		if err := queueService.ShutDown(timeoutCtx); nil != err {
			_logger.Warn("violence shutdown by signal: " + err.Error())
		} else {
			_logger.Info("graceful shutdown by signal")
		}

		// closer close
		close(idleCloser)
	}()

	// start worker daemon
	if err := queueService.Start(); nil != err && !errors.Is(err, queue.ErrQueueClosed) {
		_logger.Info("queue started failed: " + err.Error())
		close(idleCloser)
	} else {
		_logger.Info("queue worker started")
	}

	// test dispatch task after daemon started 10 second
	time.AfterFunc(10*time.Second, func() {
		// test_task
		fmt.Printf("1.queue len is %d\n", queueService.Size(&tasks.TestTask{}))
		err := queueService.Dispatch(&tasks.TestTask{}, "dispatch task")
		if err != nil {
			fmt.Printf("dispatch taks error: %s", err.Error())
		}

		fmt.Printf("2.queue len is %d\n", queueService.Size(&tasks.TestTask{}))
		err = queueService.Delay(&tasks.TestTask{}, "delay task", 10*time.Second)
		if err != nil {
			fmt.Printf("delay taks error: %s", err.Error())
		}

		fmt.Printf("3.queue len is %d\n", queueService.Size(&tasks.TestTask{}))
		err = queueService.DelayAt(&tasks.TestTask{}, "delayAt task", time.Now().Add(5*time.Second))
		if err != nil {
			fmt.Printf("delayAt taks error: %s", err.Error())
		}

		// test_timeout
		fmt.Printf("1.queue len is %d\n", queueService.Size(&tasks.TestTask{}))
		err = queueService.Dispatch(&tasks.TestTimeout{}, "dispatch task")
		if err != nil {
			fmt.Printf("dispatch taks error: %s", err.Error())
		}

		//fmt.Printf("2.queue len is %d\n", queueService.Size(&tasks.TestTask{}))
		//err = queueService.Delay(&tasks.TestTimeout{}, "delay task", 10 * time.Second)
		//if err != nil {
		//	fmt.Printf("delay taks error: %s", err.Error())
		//}
		//
		//fmt.Printf("3.queue len is %d\n", queueService.Size(&tasks.TestTask{}))
		//err = queueService.DelayAt(&tasks.TestTimeout{}, "delayAt task", time.Now().Add(5 * time.Second))
		//if err != nil {
		//	fmt.Printf("delayAt taks error: %s", err.Error())
		//}
	})

	<-idleCloser
	_logger.Info("queue worker quit, daemon exited")
}
