package tasks

import (
	"context"
	"fmt"
	"github.com/jjonline/go-mod-library/queue"
	"time"
)

type TestTask struct {
	queue.DefaultTaskSetting
}

func (t TestTask) Name() string {
	return "test_task"
}

func (t TestTask) Execute(ctx context.Context, job *queue.RawBody) error {
	fmt.Println(job.ID)
	time.Sleep(12 * time.Second)
	return fmt.Errorf("test error")
}



