package main

import (
	"github.com/jjonline/go-lib-backend/crontab"
	"github.com/jjonline/go-lib-backend/example/crontab/console"
	"github.com/jjonline/go-lib-backend/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// init
	log := logger.New("debug", "stderr")
	siCrontab := crontab.New(log.Zap.With(zap.String("module", "crontab")))

	// register
	siCrontab.Register("0 */1 * * * *", &console.TestCommandOk{})
	siCrontab.Register("0 */1 * * * *", &console.TestCommandFail{})
	siCrontab.Register("0 */1 * * * *", &console.TestCommandPanic{})

	// 接收退出信号
	quitChan := make(chan os.Signal)
	signal.Notify(
		quitChan,
		syscall.SIGINT,  // 用户发送INTR字符(Ctrl+C)触发
		syscall.SIGTERM, // 结束程序
		syscall.SIGHUP,  // 终端控制进程结束(终端连接断开)
		syscall.SIGQUIT, // 用户发送QUIT字符(Ctrl+/)触发
	)

	// region main主进程阻塞channel
	idleCloser := make(chan struct{})
	// endregion
	// start
	siCrontab.Start()
	select {
	case <-quitChan:
		siCrontab.Shutdown()
		close(idleCloser)
	}
	<-idleCloser
	return
}
