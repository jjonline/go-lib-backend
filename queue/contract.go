/*
 * 定义队列契约
 * @Time   : 2021/1/8 上午9:42
 * @Email  : jjonline@jjonline.cn
 */
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const (
	// shutdownPollIntervalMax 优雅关闭进程最大重复尝试间隔时长
	shutdownPollIntervalMax = 500 * time.Millisecond
	// DefaultMaxExecuteDuration job任务执行时长极限预警值：15分钟
	DefaultMaxExecuteDuration = 900 * time.Second
)

var (
	// ErrQueueClosed 队列处于优雅关闭或关闭状态错误
	ErrQueueClosed = errors.New("queue.error.queue.closed")
	// ErrMaxAttemptsExceeded 尝试执行次数超限
	ErrMaxAttemptsExceeded = errors.New("queue.max.execute.attempts")
	// ErrAbortForWaitingPrevJobFinish 等待上一次任务执行结束退出
	ErrAbortForWaitingPrevJobFinish = errors.New("queue.abort.for.waiting.prev.job.finish")
)

// 任务输出相关文案变量统一定义：便于日志追踪
var (
	textJobProcessing = "queue.job.processing"   // job开始执行标记文案
	textJobProcessed  = "queue.job.processed"    // job已执行成功标记文案
	textJobFailed     = "queue.job.failed"       // job已执行失败标记文案<任务类返回了error>
	textJobTooLong    = "queue.execute.too.long" // job多次尝试执行检查距离上次执行时间差已经大于 DefaultMaxExecuteDuration
	textJobFailedLog  = "queue.failed.log"       // job执行失败标记文案
)

// region queue队列抽象

// QueueIFace 基于不同技术栈的队列实现契约
type QueueIFace interface {
	// Size 获取当前队列长度方法
	// @param queue 队列的名称
	Size(queue string) (size int64)
	// Push 投递一条任务到队列方法
	// @param queue 队列的名称
	// @param payload 投递进队列的参数负载
	Push(queue string, payload interface{}) (err error)
	// Later 投递一条指定延长时长的延迟任务到队列的方法
	// @param queue 延迟队列的名称
	// @param durationTo 相对于投递任务时刻延迟的时长
	// @param payload 投递进队列的多个参数负载
	Later(queue string, durationTo time.Duration, payload interface{}) (err error)
	// LaterAt 投递一条指定执行时间的延迟任务到队列的方法
	// @param queue 延迟队列的名称
	// @param timeAt 延迟执行的时刻
	// @param payload 投递进队列的多个参数负载
	LaterAt(queue string, timeAt time.Time, payload interface{}) (err error)
	// Pop 从队尾取出一条任务的方法
	// @param queue 队列的名称
	Pop(queue string) (job JobIFace, exist bool)
	// SetConnection 设置队列底层连接器
	// @param connection 底层连接器实例
	SetConnection(connection interface{}) (err error)
	// GetConnection 获取队列底层连接器
	GetConnection() (connection interface{}, err error)
}

// endregion

// region job任务抽象

// JobIFace 基于不同技术栈的队列任务Job实现契约
type JobIFace interface {
	Release(delay int64) (err error) // 释放任务：将任务重新放入队列
	Delete() (err error)             // 删除任务：任务不再执行
	IsDeleted() (deleted bool)       // 检查任务是否已删除
	IsReleased() (released bool)     // 检查任务是否已释放
	Attempts() (attempt int64)       // 获取任务已尝试执行过的次数
	PopTime() (time time.Time)       // 获取任务首次被pop取出的时刻
	Timeout() (time time.Duration)   // 任务超时时长
	TimeoutAt() (time time.Time)     // 任务执行超时的时刻
	HasFailed() (hasFail bool)       // 检测当前job任务执行是否出现了错误
	MarkAsFailed()                   // 设置当前job任务执行出现了错误
	Failed(err error)                // 设置任务执行失败
	Queue() (queue QueueIFace)       // 获取job任务所属队列queue句柄
	GetName() (queueName string)     // 获取job所属队列名称
	Payload() (payload *Payload)     // 获取任务执行参数payload
}

// endregion

// region 定义任务传参实体RawBody

// RawBody 队列execute执行时传递给执行方法的参数Raw结构：job任务参数的包装器
type RawBody struct {
	queue   string // 队列名
	payload []byte // 调度队列塞入的数据体
	ID      string // 队列内部唯一标识符ID
}

// Int 任务数据转int
func (rawBody *RawBody) Int() int {
	i, _ := strconv.Atoi(string(rawBody.payload))
	return i
}

// String 任务数据转string
func (rawBody *RawBody) String() string {
	return string(rawBody.payload)
}

// Bytes 任务数据转[]byte
func (rawBody *RawBody) Bytes() []byte {
	return rawBody.payload
}

// Int64 任务数据转int64
func (rawBody *RawBody) Int64() int64 {
	i64, _ := strconv.ParseInt(string(rawBody.payload), 10, 64)
	return i64
}

// Unmarshal 任务数据Unmarshal为投递调度任务时的结构类型
// result 具体类型的指针引用变量，转换成功将自动填充
// 转换成功填充result返回nil，转换失败时返回error
func (rawBody *RawBody) Unmarshal(result interface{}) error {
	return json.Unmarshal(rawBody.payload, result)
}

// endregion

// region 队列任务job单元存储struct && 失败任务处理器方法签名定义

// Payload 存储于队列中的job任务结构
type Payload struct {
	Name          string `json:"Name"`          // 队列名称
	ID            string `json:"ID"`            // 任务ID
	MaxTries      int64  `json:"MaxTries"`      // 任务最大尝试次数，默认1
	RetryInterval int64  `json:"RetryInterval"` // 当任务最大允许尝试次数大于0时，下次尝试之前的间隔时长，单位：秒
	Attempts      int64  `json:"Attempts"`      // 任务已被尝试执行的的次数
	Payload       []byte `json:"Payload"`       // 任务参数比特字面量，可decode成具体job被execute时的类型
	PopTime       int64  `json:"PopTime"`       // 任务首次被取出执行的时间戳
	Timeout       int64  `json:"Timeout"`       // 任务最大执行超时时长，单位：秒
	TimeoutAt     int64  `json:"TimeoutAt"`     // 任务超时时刻时间戳
}

// RawBody PayLoad结构体获取载体实体
func (payload *Payload) RawBody() *RawBody {
	return &RawBody{queue: payload.Name, ID: payload.ID, payload: payload.Payload}
}

// FailedJobHandler 失败任务记录|处理回调方法
// @param *Payload 失败job的对象信息
// @param error job任务失败的error报错信息
type FailedJobHandler func(payload *Payload, err error) error

// endregion

// region 任务类契约 && 任务类默认设置嵌入结构体

// TaskIFace 定义队列Job任务执行逻辑的契约(队列任务执行类)
type TaskIFace interface {
	MaxTries() int64                                 // 定义队列任务最大尝试次数：任务执行的最大尝试次数
	RetryInterval() int64                            // 定义队列任务最大尝试间隔：当任务执行失败后再次尝试执行的间隔时长，单位：秒
	Timeout() time.Duration                          // 定义队列超时方法：返回超时时长
	Name() string                                    // 定义队列名称方法：返回队列名称
	Execute(ctx context.Context, job *RawBody) error // 定义队列任务执行时的方法：执行成功返回nil，执行失败返回error
}

// DefaultTaskSetting 默认task设置struct：实现默认的最大尝试次数、尝试间隔时长、最大执行时长
type DefaultTaskSetting struct{}

// MaxTries 默认最大尝试次数1，即投递的任务仅执行1次
func (task *DefaultTaskSetting) MaxTries() int64 {
	return 1
}

// RetryInterval 当任务执行失败后再次尝试执行的间隔时长，默认立即重试，即间隔时长为0秒
func (task *DefaultTaskSetting) RetryInterval() int64 {
	return 0
}

// Timeout 任务最大执行超时时长：默认超时时长为900秒
func (task *DefaultTaskSetting) Timeout() time.Duration {
	return DefaultMaxExecuteDuration
}

// DefaultTaskSettingWithoutTimeout 默认task设置struct：实现默认的最大尝试次数、尝试间隔时长、最大执行时长
type DefaultTaskSettingWithoutTimeout struct{}

// MaxTries 默认最大尝试次数1，即投递的任务仅执行1次
func (task *DefaultTaskSettingWithoutTimeout) MaxTries() int64 {
	return 1
}

// RetryInterval 当任务执行失败后再次尝试执行的间隔时长，默认立即重试，即间隔时长为0秒
func (task *DefaultTaskSettingWithoutTimeout) RetryInterval() int64 {
	return 0
}

// jobProperty 公共的job实现类内部属性
type jobProperty struct {
	handler    QueueIFace    // 所属队列实现hand
	name       string        // 队列名字
	job        string        // job内部存储实体
	reserved   string        // 已标记执行中job内部存储实体
	payload    *Payload      // job任务payload
	isReleased bool          // 是否已释放标记
	isDeleted  bool          // 是否已删除标记
	hasFailed  bool          // 是否已失败标记
	popTime    time.Time     // 任务被pop取出的时刻（等级于开始执行时刻）
	timeout    time.Duration // 任务超时时长
	timeoutAt  time.Time     // 任务执行超时的时刻
}

// endregion
