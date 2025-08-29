package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

/*
 * @Time   : 2025-08-27 09:34:31
 * @Desc   : 基于MySQL实现的Job
 */

type JobMySQL struct {
	basic      queueBasic  // 引入基础公用方法
	db         *sql.DB     // MySQL数据库连接
	tableID    int64       // 数据库表记录ID
	lock       sync.Mutex  // 防幻读锁
	mysqlQueue *mysqlQueue // MySQL队列引用，用于获取表前缀
	jobProperty
}

// Release 释放任务job：job重新再试--清除reserved_at标记，设置新的available_at延迟时间
func (job *JobMySQL) Release(delay int64) (err error) {
	job.lock.Lock()
	defer job.lock.Unlock()

	job.isReleased = true

	// 计算新的可用时间
	availableAt := time.Now().Add(time.Duration(delay) * time.Second).Unix()

	// 更新数据库记录，清除reserved_at并设置新的available_at
	query := `UPDATE ` + job.mysqlQueue.getJobsTableName() + ` SET reserved_at = NULL, available_at = ? WHERE id = ?`
	_, err = job.db.Exec(query, availableAt, job.tableID)

	return err
}

// Delete 删除任务job：任务不再执行--从数据库删除记录
func (job *JobMySQL) Delete() (err error) {
	job.lock.Lock()
	defer job.lock.Unlock()
	job.isDeleted = true

	// 从数据库删除任务记录
	query := `DELETE FROM ` + job.mysqlQueue.getJobsTableName() + ` WHERE id = ?`
	_, err = job.db.Exec(query, job.tableID)

	return err
}

func (job *JobMySQL) IsDeleted() (deleted bool) {
	job.lock.Lock()
	defer job.lock.Unlock()
	return job.isDeleted
}

func (job *JobMySQL) IsReleased() (released bool) {
	job.lock.Lock()
	defer job.lock.Unlock()
	return job.isReleased
}

// Attempts 获取当前job已被尝试执行的次数
func (job *JobMySQL) Attempts() (attempt int64) {
	return job.payload.Attempts
}

// PopTime 任务job首次被执行的时刻
func (job *JobMySQL) PopTime() (time time.Time) {
	return job.popTime
}

// Timeout 任务超时时长
func (job *JobMySQL) Timeout() (time time.Duration) {
	return job.jobProperty.timeout
}

// TimeoutAt 任务job执行超时的时刻
func (job *JobMySQL) TimeoutAt() (time time.Time) {
	return job.jobProperty.timeoutAt
}

func (job *JobMySQL) HasFailed() (hasFail bool) {
	job.lock.Lock()
	defer job.lock.Unlock()
	return job.hasFailed
}

func (job *JobMySQL) MarkAsFailed() {
	job.lock.Lock()
	defer job.lock.Unlock()
	job.hasFailed = true
}

func (job *JobMySQL) Failed(err error) {
	// 可选：将失败任务记录到失败表中
	job.recordFailedJob(err)
}

func (job *JobMySQL) GetName() (queueName string) {
	return job.name
}

func (job *JobMySQL) Queue() (queue QueueIFace) {
	return job.handler
}

func (job *JobMySQL) Payload() (payload *Payload) {
	return job.payload
}

// recordFailedJob 记录失败任务到失败表（可选实现）
func (job *JobMySQL) recordFailedJob(err error) {
	// 记录失败任务到failed jobs表
	failedAt := time.Now().Unix()
	payloadBytes, _ := json.Marshal(job.payload)

	query := fmt.Sprintf(`INSERT INTO %s (queue_name, payload, exception, failed_at) VALUES (?, ?, ?, ?)`, job.mysqlQueue.getFailedJobsTableName())
	_, dbErr := job.db.Exec(query, job.name, string(payloadBytes), err.Error(), failedAt)

	if dbErr != nil {
		// 如果记录失败任务失败，这里可以记录日志但不抛出错误
		fmt.Printf("Failed to record failed job: %v\n", dbErr)
	}
}
