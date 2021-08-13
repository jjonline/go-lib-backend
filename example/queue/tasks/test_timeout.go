package tasks

import (
	"context"
	"fmt"
	"github.com/jjonline/go-lib-backend/queue"
	"time"
)

type TestTimeout struct {
	queue.DefaultTaskSettingWithoutTimeout
}

// Timeout 任务最大执行超时时长：默认超时时长为900秒
func (task *TestTimeout) Timeout() time.Duration {
	return 5 * time.Second
}

func (t TestTimeout) Name() string {
	return "test_timeout"
}

func (t TestTimeout) Execute(ctx context.Context, job *queue.RawBody) error {
	select {
	case <- ctx.Done():
		return ctx.Err()
	default:
		time.Sleep(6 * time.Second)
		fmt.Println("job execute finished")
		return nil
	}
}
