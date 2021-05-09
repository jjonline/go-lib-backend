package tasks

import (
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

func (t TestTask) Execute(job *queue.RawBody) error {
	fmt.Println(job.ID)
	time.Sleep(12 * time.Second)
	return fmt.Errorf("test error")
}



