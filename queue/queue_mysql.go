package queue

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// 基于MySQL实现队列机制：
// 一、原理
//    使用MySQL表实现队列机制，通过available_at字段控制延迟执行，reserved_at字段标记任务是否被消费者获取
// 二、producer
// 	  实时队列：往queue_jobs表插入数据，available_at为当前时间戳
//    延时队列：往queue_jobs表插入数据，available_at为延迟执行时间戳
// 三、consumer/worker步骤
//    step1、查询available_at小于等于当前时间戳且reserved_at为NULL的任务
//    step2、更新reserved_at字段为超时时间戳，并增加attempts计数
//    step3、执行任务，成功删除记录，失败根据重试策略处理
// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// mysqlQueue 基于MySQL实现的队列
// implement QueueIFace
type mysqlQueue struct {
	queueBasic             // 队列基础可公用方法
	connection  *sql.DB    // MySQL数据库连接
	lock        sync.Mutex // 并发锁
	tablePrefix string     // 表前缀
}

// getJobsTableName 获取队列任务表名
func (m *mysqlQueue) getJobsTableName() string {
	if m.tablePrefix != "" {
		return m.tablePrefix + "queue_jobs"
	}
	return "queue_jobs"
}

// getFailedJobsTableName 获取失败任务表名
func (m *mysqlQueue) getFailedJobsTableName() string {
	if m.tablePrefix != "" {
		return m.tablePrefix + "queue_failed_jobs"
	}
	return "queue_failed_jobs"
}

// Size 获取队列长度
func (m *mysqlQueue) Size(queue string) (size int64) {
	var count int64
	query := `SELECT COUNT(*) FROM ` + m.getJobsTableName() + ` WHERE queue_name = ? AND (reserved_at IS NULL OR reserved_at <= ?)`
	now := time.Now().Unix()

	err := m.connection.QueryRow(query, queue, now).Scan(&count)
	if err != nil {
		return 0
	}

	return count
}

// Push 投递一条任务到队列
func (m *mysqlQueue) Push(queue string, payload interface{}) (err error) {
	now := time.Now().Unix()
	query := `INSERT INTO ` + m.getJobsTableName() + ` (queue_name, payload, attempts, available_at, created_at) VALUES (?, ?, 0, ?, ?)`

	_, err = m.connection.Exec(query, queue, string(payload.([]byte)), now, now)
	return err
}

// Later 延迟指定时长后执行的延迟任务
func (m *mysqlQueue) Later(queue string, durationTo time.Duration, payload interface{}) (err error) {
	return m.LaterAt(queue, time.Now().Add(durationTo), payload)
}

// LaterAt 指定时刻执行的延时任务
func (m *mysqlQueue) LaterAt(queue string, timeAt time.Time, payload interface{}) (err error) {
	now := time.Now().Unix()
	availableAt := timeAt.Unix()
	query := `INSERT INTO ` + m.getJobsTableName() + ` (queue_name, payload, attempts, available_at, created_at) VALUES (?, ?, 0, ?, ?)`

	_, err = m.connection.Exec(query, queue, payload.([]byte), availableAt, now)
	return err
}

// Pop 取出弹出一条待执行的任务
func (m *mysqlQueue) Pop(queue string) (job JobIFace, exist bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 1. 延时任务结束延时，标记可被执行
	// 2. 已达到超时仍然未release释放的reserved保留任务，标记可再次被执行
	// 3. 读取一条需立即执行的数据并标记为reserved状态
	// 4. 执行成功：移除reserved状态的数据；执行失败：移除reserved状态的数据并触发Failed方法

	var (
		now                = time.Now()
		nowUnix            = now.Unix()
		transitionHasError error
	)

	// MySQL事务内执行
	tx, err := m.connection.Begin()
	if err != nil {
		return nil, false
	}

	// defer rollback
	defer func() {
		if transitionHasError != nil {
			_ = tx.Rollback()
		}
	}()

	// step2: 释放超时的reserved任务（相当于Redis的migrated expired reserved jobs）
	releaseQuery := `UPDATE ` + m.getJobsTableName() + ` SET reserved_at = NULL WHERE queue_name = ? AND reserved_at IS NOT NULL AND reserved_at <= ?`
	if _, err = tx.Exec(releaseQuery, queue, nowUnix); err != nil {
		transitionHasError = err
		return nil, false
	}

	// step3: 查询可用任务
	var (
		id          int64
		payloadStr  string
		attempts    int64
		selectQuery = `SELECT id, payload, attempts FROM ` + m.getJobsTableName() + ` WHERE queue_name = ? AND available_at <= ? AND reserved_at IS NULL ORDER BY id ASC LIMIT 1 FOR UPDATE`
	)
	err = tx.QueryRow(selectQuery, queue, nowUnix).Scan(&id, &payloadStr, &attempts)
	if err != nil {
		transitionHasError = err
		return nil, false
	}

	// step4： 解析payload，构造必要参数
	var payloadData Payload
	if err = json.Unmarshal([]byte(payloadStr), &payloadData); err != nil {
		transitionHasError = err
		return nil, false
	}

	// step5： 更新查询出的任务为reserved状态，并增加attempts
	var (
		reservedAt  = now.Add(time.Duration(payloadData.Timeout) * time.Second).Unix()
		updateQuery = `UPDATE ` + m.getJobsTableName() + ` SET reserved_at = ?, attempts = attempts + 1 WHERE id = ?`
	)
	if _, err = tx.Exec(updateQuery, reservedAt, id); err != nil {
		transitionHasError = err
		return nil, false
	}

	// 设置首次被取出时间
	if payloadData.PopTime <= 0 {
		payloadData.PopTime = nowUnix // 更新数据库中的payload
		if updatedPayload, err1 := json.Marshal(payloadData); err1 == nil {
			updatePayloadQuery := `UPDATE ` + m.getJobsTableName() + ` SET payload = ? WHERE id = ?`
			if _, err2 := tx.Exec(updatePayloadQuery, string(updatedPayload), id); err2 != nil {
				transitionHasError = err2
				return nil, false
			}
		} else {
			transitionHasError = err1
			return nil, false
		}
	}

	if err = tx.Commit(); err != nil {
		transitionHasError = err
		return nil, false
	}

	// 增加尝试次数
	payloadData.Attempts = attempts + 1

	return &JobMySQL{
		db:         m.connection,
		lock:       sync.Mutex{},
		tableID:    id,
		mysqlQueue: m,
		jobProperty: jobProperty{
			handler:    m,
			name:       queue,
			job:        payloadStr,
			reserved:   "",
			payload:    &payloadData,
			isReleased: false,
			isDeleted:  false,
			hasFailed:  false,
			popTime:    time.Unix(payloadData.PopTime, 0),
			timeout:    time.Duration(payloadData.Timeout) * time.Second,
			timeoutAt:  now.Add(time.Duration(payloadData.Timeout) * time.Second),
		},
	}, true
}

// SetConnection 设置MySQL队列的连接器：sql.DB实例指针
func (m *mysqlQueue) SetConnection(connection interface{}) (err error) {
	db, ok := connection.(*sql.DB)
	if !ok {
		return errors.New("connection must be *sql.DB type")
	}

	m.connection = db

	// 测试连接
	if err := m.connection.Ping(); err != nil {
		return errors.New("mysql connection test failed: " + err.Error())
	}

	return nil
}

// GetConnection 获取MySQL队列的连接器：sql.DB实例指针（interface）使用前需显式转换
// example:
//
//	conn, _ := m.GetConnection()
//	db := conn.(*sql.DB)
//	db.Exec("SELECT 1")
func (m *mysqlQueue) GetConnection() (connection interface{}, err error) {
	if m.connection == nil {
		return nil, errors.New("null pointer connection instance")
	}

	return m.connection, nil
}
