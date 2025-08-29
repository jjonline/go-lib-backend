package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jjonline/go-lib-backend/queue"
)

// SimpleLogger 简单的日志记录器实现
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string, keyValue ...any) {
	slog.Debug(msg, keyValue...)
}

func (l *SimpleLogger) Info(msg string, keyValue ...any) {
	slog.Info(msg, keyValue...)
}

func (l *SimpleLogger) Warn(msg string, keyValue ...any) {
	slog.Warn(msg, keyValue...)
}

func (l *SimpleLogger) Error(msg string, keyValue ...any) {
	slog.Error(msg, keyValue...)
}

// EmailTask 示例任务类
type EmailTask struct {
	queue.DefaultTaskSetting
}

func (t EmailTask) Name() string {
	return "email_task"
}

func (t EmailTask) Execute(ctx context.Context, job *queue.RawBody) error {
	// 解析任务参数
	var emailData struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := job.Unmarshal(&emailData); err != nil {
		return fmt.Errorf("failed to unmarshal email data: %w", err)
	}

	// 模拟发送邮件
	fmt.Printf("Sending email to: %s, Subject: %s, Body: %s\n", emailData.To, emailData.Subject, emailData.Body)

	// 模拟处理时间
	time.Sleep(1 * time.Second)

	fmt.Printf("Email sent successfully! Job ID: %s\n", job.ID)

	panic("test panic")

	return fmt.Errorf("email sent successfully! Job ID: %s", job.ID)
}

func (t EmailTask) Remark() string {
	return "发送邮件任务"
}

func main() {
	// todo 数据库连接配置，请修改为可用的mysql
	dsn := "root:password@tcp(localhost:3306)/test_queue?charset=utf8mb4&parseTime=True&loc=Local"

	// 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// set slog default handler
	handler := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	slog.SetDefault(handler)

	// 初始化队列
	logger := &SimpleLogger{}
	queueService := queue.New(
		queue.MySQL,                       // 使用MySQL驱动
		db,                                // 数据库连接
		logger,                            // 日志记录器
		queue.Config{MaxConcurrency: 100}, // 最大并发消费者数量
	)

	// 注册任务类
	emailTask := &EmailTask{}
	if err := queueService.BootstrapOne(emailTask); err != nil {
		log.Fatalf("Failed to bootstrap task: %v", err)
	}

	// 设置失败任务处理器
	queueService.SetFailedJobHandler(func(payload *queue.Payload, err error) error {
		fmt.Printf("Job failed: %s, Error: %s\n", payload.Name, err.Error())
		return nil
	})

	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Printf("%#v\n", queueService.GetStatistics())
			_ = queueService.AutoScaleWorkers()
		}
	}()

	// 投递一些任务
	fmt.Println("Dispatching jobs...")

	// 投递立即执行的任务
	go func() {
		time.Sleep(5 * time.Second)
		for i := 1; i <= 100; i++ {
			emailData := map[string]interface{}{
				"to":      fmt.Sprintf("user%d@example.com", i),
				"subject": fmt.Sprintf("测试邮件 %d", i),
				"body":    fmt.Sprintf("这是第 %d 封测试邮件", i),
			}

			if err := queueService.Dispatch(emailTask, emailData); err != nil {
				fmt.Printf("Failed to dispatch job %d: %v\n", i, err)
			}
		}
	}()

	// 投递延迟任务
	delayEmailData := map[string]interface{}{
		"to":      "delayed@example.com",
		"subject": "延迟邮件",
		"body":    "这是一封延迟5秒发送的邮件",
	}

	if err := queueService.Delay(emailTask, delayEmailData, 5*time.Second); err != nil {
		fmt.Printf("Failed to dispatch delayed job: %v\n", err)
	}

	// 调试：检查队列状态
	fmt.Println("\n=== Queue Status Before Starting Consumer ===")

	// 启动队列消费者
	fmt.Println("Starting queue consumer...")
	if err := queueService.Start(); err != nil {
		log.Fatalf("Failed to start queue: %v", err)
	}

	// 等待中断信号
	fmt.Println("Queue is running. Press Ctrl+C to stop.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭队列
	fmt.Println("Shutting down queue...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := queueService.ShutDown(ctx); err != nil {
		fmt.Printf("Queue shutdown error: %v\n", err)
	} else {
		fmt.Println("Queue shutdown successfully")
		time.Sleep(3 * time.Second)
	}
}
