package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-stack/stack"
	"github.com/shirou/gopsutil/v4/mem"
)

// *************************************************
// 队列管理者
// 1、实际维护已注册的任务类
// 2、维护管理工作进程worker
// 3、队列相关管控功能实现：启动、优雅停止、协程并发调度等
// *************************************************

const (
	general                   = "general"              // 通用名称
	jitterBase                = 450 * time.Millisecond // looper最小为450毫秒间隔，最大为1000毫秒间隔
	memoryMaxPercentThreshold = 90                     // 系统内存使用率阈值后停止自动扩容
)

type atomicBool int32

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }
func (b *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(b), 0) }

// manager 队列管理者，队列的调度执行和管理
type manager struct {
	queue            QueueIFace               // 队列底层实现实例
	channel          chan JobIFace            // 任务类执行job的通道chan
	logger           Logger                   // 实现 Logger 接口的结构体实例的指针对象
	config           Config                   // 队列配置
	concurrent       int64                    // 当前并发worker数
	tasks            map[string]TaskIFace     // 队列名与任务类实例映射map，interface无需显式指定执指针类型，但实际传参需指针类型
	failedJobHandler FailedJobHandler         // 失败任务[最大尝试次数后仍然尝试失败（Execute返回了Error 或 执行导致panic）的任务]处理器
	lock             sync.Mutex               // 并发锁
	doneChan         chan struct{}            // 关闭队列的信号控制chan
	inShutdown       atomicBool               // 原子态标记：是否处于优雅关闭状态中
	isChannelClosed  atomicBool               // 原子态标记：looper与worker之间channel是否已关闭，多个looper争抢关闭channel
	inWorkingMap     sync.Map                 // map[string]int64  当前正work中的jobID与workerID映射map
	workerStatus     map[int64]*atomicBool    // worker工作进程状态标记map
	workerChannel    map[int64]chan struct{}  // worker停止信号通道映射map
	jitter           map[string]time.Duration // 循环器抖动间隔，key为task或general，value为对应looper的循环间隔
	allowTasks       map[string]struct{}      // 指定可以运行的队列
	excludeTasks     map[string]struct{}      // 指定不可运行的队列
	realTasksNum     int64                    // 可以运行的task数（综合计算task、allowTasks、canExecuteTask）
	nextWorkerID     int64                    // 下一个worker ID
}

// newManager 实例化一个manager
// @param queue    队列实现底层实例指针
// @param logger   实现 Logger 接口的结构体实例的指针对象
// @param config   配置
func newManager(queue QueueIFace, logger Logger, config Config) *manager {
	return &manager{
		queue:         queue,
		channel:       make(chan JobIFace), // no buffer channel, execute when worker received
		logger:        logger,
		config:        config,
		tasks:         make(map[string]TaskIFace),
		workerStatus:  make(map[int64]*atomicBool),
		workerChannel: make(map[int64]chan struct{}),
		inWorkingMap:  sync.Map{},
		lock:          sync.Mutex{},
		jitter:        make(map[string]time.Duration),
		allowTasks:    make(map[string]struct{}),
		excludeTasks:  make(map[string]struct{}),
		realTasksNum:  0,
		nextWorkerID:  0,
	}
}

// bootstrapOne 脚手架辅助载入注册一个任务类
func (m *manager) bootstrapOne(task TaskIFace) error {
	m.lock.Lock()

	// log
	m.logger.Debug(
		"bootstrap",
		"name",
		task.Name(),
		"max_tries",
		IFaceToString(task.MaxTries()),
		"retry_interval",
		IFaceToString(task.RetryInterval()),
	)

	m.tasks[task.Name()] = task
	m.lock.Unlock()

	return nil
}

// bootstrap 脚手架辅助载入注册多个任务类
func (m *manager) bootstrap(tasks []TaskIFace) (err error) {
	for _, job := range tasks {
		if err = m.bootstrapOne(job); nil != err {
			return err
		}
	}
	return nil
}

// start 启动队列进程工作者
func (m *manager) start() (err error) {
	// 队列处于关闭中状态时启动直接返回Err
	if m.shuttingDown() {
		return ErrQueueClosed
	}

	// ① 启动通用looper
	go m.startGeneralLooper()

	// ② 启动task专用looper
	m.startDedicatedLooper()

	// ③ 启动task数量的worker
	m.lock.Lock()
	m.startSingleWorker() // for general +1 worker
	for name := range m.tasks {
		// 检查任务是否可以运行
		if !m.allowRun(name) {
			continue
		}
		m.realTasksNum++ // 记录真实运行运行的task数
		m.startSingleWorker()
	}
	m.lock.Unlock()

	// ④ 启动自动扩缩容检测器
	go m.startAutoScaleMonitor()

	return err
}

// startGeneralLooper 启动通用looper，用于loop所有task
func (m *manager) startGeneralLooper() {
	for {
		select {
		case <-m.getDoneChan():
			m.logger.Info("shutdown, queue general looper exited")
			m.closeChannel() // close job chan
			return
		default:
			m.generalLooper() // continue loop all queue jobs
		}
	}
}

// generalLooper 通用轮询 && 速率控制所有队列的looper
func (m *manager) generalLooper() {
	// map的range是无序的，无需再随机pop队列
	// range本身就是随机的
	needSleep := true

	for name := range m.tasks {
		//检查任务是否可以运行
		if !m.allowRun(name) {
			continue
		}

		// chan关闭则退出
		if m.isChannelClosed.isSet() {
			return
		}

		if job, exist := m.queue.Pop(name); exist {
			m.channel <- job // push job to worker for control process
			needSleep = false
		}
	}

	// 所有队列都没job任务 looper随机休眠
	if needSleep {
		m.logger.Debug("no job pop, sleep for a while", "task", general)

		time.Sleep(m.looperJitter(general))
	}
}

// startDedicatedLooper 启动专用looper，用于loop指定task
func (m *manager) startDedicatedLooper() {
	for name := range m.tasks {
		// 检查任务是否可以运行
		if !m.allowRun(name) {
			continue
		}

		// 启动专用looper
		go m.dedicatedLooper(name)
	}
}

// dedicatedLooper 专用轮询 && 速率控制所有队列的looper
func (m *manager) dedicatedLooper(name string) {
	for {
		select {
		case <-m.getDoneChan():
			m.logger.Info("shutdown, queue dedicated looper exited", "task_looper", name)
			return
		default:
			// chan关闭则退出
			if m.isChannelClosed.isSet() {
				return
			}

			// map的range是无序的，无需再随机pop队列
			// range本身就是随机的
			needSleep := true

			if job, exist := m.queue.Pop(name); exist {
				m.channel <- job // push job to worker for control process
				needSleep = false
			}

			// 队列暂无job任务 looper随机休眠
			if needSleep {
				m.logger.Debug("no job pop, sleep for a while", "task_looper", name)

				time.Sleep(m.looperJitter(name))
			}
		}
	}
}

// startAutoScaleMonitor 启动自动扩缩容监测器
func (m *manager) startAutoScaleMonitor() {
	if !m.config.AutoScale {
		return
	}

	ticker := time.NewTicker(m.config.AutoScaleInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.logger.Debug("start.autoScale.monitor")
		_ = m.autoScaleWorkers()
	}
}

// 检查任务是否可以运行
func (m *manager) allowRun(jobName string) bool {
	if _, ok := m.allowTasks[jobName]; len(m.allowTasks) > 0 && !ok {
		return false
	}
	if _, ok := m.excludeTasks[jobName]; len(m.excludeTasks) > 0 && ok {
		return false
	}
	return true
}

// startSingleWorker 启动单个worker进程（需要在持有锁的情况下调用）
func (m *manager) startSingleWorker() {
	workerID := m.nextWorkerID
	m.nextWorkerID++

	// 创建worker停止信号通道
	stopChan := make(chan struct{})
	m.workerChannel[workerID] = stopChan

	// 初始化worker状态
	m.workerStatus[workerID] = new(atomicBool)

	// 启动worker goroutine
	go m.startWorker(workerID, stopChan)
}

// startWorker 启动队列进程工作者
func (m *manager) startWorker(workerID int64, stopChan chan struct{}) {
	defer func() {
		// 清理worker相关资源
		m.lock.Lock()
		delete(m.workerStatus, workerID)
		delete(m.workerChannel, workerID)
		m.lock.Unlock()

		m.logger.Info(fmt.Sprintf("queue worker-%d exited", workerID), "worker_id", IFaceToString(workerID))
	}()

	// started logger
	m.logger.Info(fmt.Sprintf("queue worker-%d started", workerID), "worker_id", IFaceToString(workerID))

	// 阻塞消费job chan或等待停止信号
	for {
		select {
		case job, ok := <-m.channel:
			if !ok {
				// channel已关闭，退出worker
				return
			}
			m.runJob(job, workerID) // process run job
		case <-stopChan:
			// 收到停止信号，退出worker
			return
		}
	}
}

// runJob 执行队列job，超时控制 && 尝试次数控制，执行结果控制
func (m *manager) runJob(job JobIFace, workerID int64) {
	// set worker is true
	m.setWorkerStatus(workerID, true)

	// step1、任务类执行捕获可能的panic
	defer func() {
		// set worker execute is false
		m.setWorkerStatus(workerID, false)

		// delete in running map need to use lock
		m.inWorkingMap.Delete(job.Payload().ID)

		// recovery if panic
		if err := recover(); err != nil {
			m.logger.Error(
				"queue.execute.panic",
				"stack", stack.Trace().TrimRuntime().String(),
				"queue", job.GetName(),
				"worker_id", IFaceToString(workerID),
				"payload", IFaceToString(job.Payload()),
				"error", IFaceToString(err),
			)

			var eErr error
			switch t := err.(type) {
			case error:
				eErr = t
			default:
				eErr = fmt.Errorf("%s", t)
			}

			// panic: 检查任务尝试执行次数 & 标记失败状态
			m.markJobAsFailedIfWillExceedMaxAttempts(job, eErr)
		}
	}()

	task, ok := m.tasks[job.GetName()]
	if !ok {
		return
	}

	// step2、因为没有超时主动退出机制当任务执行超时仍在执行时标记再次延迟
	if _, exist := m.inWorkingMap.Load(job.Payload().ID); exist {
		m.logger.Warn(
			ErrAbortForWaitingPrevJobFinish.Error(),
			"queue", job.GetName(),
			"payload", IFaceToString(job.Payload()),
			"pop_time", job.PopTime().String(),
		)

		// 当前任务作为延迟任务再次投递
		// warning 当前正在执行的可能执行成功这样会导致一条任务多次被成功执行，需要任务类自主实现业务逻辑幂等
		if payload, err := json.Marshal(job.Payload()); err == nil {
			_ = job.Queue().Later(job.GetName(), time.Duration(job.Payload().RetryInterval)*time.Second, payload)
		}

		// 触发记录可能失败日志的记录，便于回溯
		m.recordFailedJob(job, ErrAbortForWaitingPrevJobFinish)

		return
	}

	// set in running map, need to be use lock
	m.inWorkingMap.Store(job.Payload().ID, workerID)

	// step3、检查任务尝试次数：超限标记任务失败后删除任务，未超限则执行
	if m.markJobAsFailedIfAlreadyExceedsMaxAttempts(job) {
		return
	}

	// step4、execute job task with timeout control
	m.logger.Info(
		textJobProcessing,
		"queue", job.GetName(),
		"worker_id", IFaceToString(workerID),
		"payload", IFaceToString(job.Payload()),
	)

	// timeout context control
	ctx, cancelFunc := context.WithTimeout(context.Background(), job.Timeout())
	defer cancelFunc()

	// 添加通信机制：done channel用于通知任务完成
	done := make(chan struct{})

	// goroutine execute task job
	go func() {
		defer func() {
			// 确保无论如何都要关闭done channel
			close(done)

			if r := recover(); r != nil {
				m.logger.Error(
					"queue.execute.panic",
					"stack", stack.Trace().TrimRuntime().String(),
					"queue", job.GetName(),
					"worker_id", IFaceToString(workerID),
					"payload", IFaceToString(job.Payload()),
					"error", IFaceToString(r),
				)

				var eErr error
				switch t := r.(type) {
				case error:
					eErr = t
				default:
					eErr = fmt.Errorf("%s", t)
				}

				// panic: 检查任务尝试执行次数 & 标记失败状态
				m.markJobAsFailedIfWillExceedMaxAttempts(job, eErr)
			}
		}()
		err := task.Execute(ctx, job.Payload().RawBody())
		if err == nil {
			// step5、任务类执行成功：删除任务即可
			m.logger.Info(
				textJobProcessed,
				"queue", job.GetName(),
				"worker_id", IFaceToString(workerID),
				"payload", IFaceToString(job.Payload()),
				"duration", IFaceToString(int64(time.Now().Sub(job.PopTime()))),
			)
			_ = job.Delete()
		} else {
			// step6、任务类执行失败：依赖重试设置执行重试or最终执行失败处理
			m.logger.Error(
				textJobFailed,
				"queue", job.GetName(),
				"worker_id", IFaceToString(workerID),
				"payload", IFaceToString(job.Payload()),
				"duration", IFaceToString(int64(time.Now().Sub(job.PopTime()))),
			)
			m.markJobAsFailedIfWillExceedMaxAttempts(job, err)
		}
	}()

	select {
	case <-done:
		// 任务已完成（成功、失败或panic），正常退出
		return
	case <-ctx.Done():
		// 任务超时，但任务可能仍在执行中
		m.logger.Warn(
			"queue.job.timeout",
			"queue", job.GetName(),
			"worker_id", IFaceToString(workerID),
			"payload", IFaceToString(job.Payload()),
			"timeout", IFaceToString(int64(job.Timeout().Seconds())),
		)
		m.markJobAsFailedIfWillExceedMaxAttempts(job, ctx.Err())
		return
	}
}

// looperJitter looper循环器间隔抖动
//
//	-- name task名或general
func (m *manager) looperJitter(name string) time.Duration {
	m.lock.Lock()
	defer m.lock.Unlock()

	// init
	if _, ok := m.jitter[name]; !ok {
		m.jitter[name] = jitterBase
	}

	m.jitter[name] = m.jitter[name] + time.Duration(rand.Intn(int(jitterBase/3)))
	if m.jitter[name] > 1*time.Second {
		m.jitter[name] = jitterBase
	}

	return m.jitter[name]
}

// markJobAsFailedIfAlreadyExceedsMaxAttempts job执行`之前`检测尝试次数是否超限
// 1、如果超限则方法体内部清理任务并返回true，表示该job需要停止执行
// 2、如果未超限则返回false
func (m *manager) markJobAsFailedIfAlreadyExceedsMaxAttempts(job JobIFace) (needSop bool) {
	// step1、执行时长检查，持续执行超过设置的超时时长则记录日志
	if time.Now().Sub(job.PopTime()) >= job.Timeout() {
		m.logger.Warn(
			textJobTooLong,
			"queue", job.GetName(),
			"payload", IFaceToString(job.Payload()),
			"pop_time", job.PopTime().String(),
		)
	}

	// step2、检查最大尝试次数
	if job.Attempts() <= job.Payload().MaxTries {
		return false
	}

	// step3、其他情况：执行job前检查就不通过，移除任务&&标记任务失败（最大尝试次数超过限制、持续执行超时、脏数据、意外中断的任务 等）
	m.failJob(job, ErrMaxAttemptsExceeded)

	return true
}

// markJobAsFailedIfWillExceedMaxAttempts job执行`之后`检测尝试次数是否超限
// 1、检查job执行是否超过基准时间以记录日志
// 2、检查job执行尝试次数
func (m *manager) markJobAsFailedIfWillExceedMaxAttempts(job JobIFace, err error) {
	if job.IsDeleted() {
		return
	}

	// step1、执行时长检查：超时记录超时日志
	if time.Now().Sub(job.PopTime()) >= job.Timeout() {
		m.logger.Warn(
			textJobTooLong,
			"queue", job.GetName(),
			"payload", IFaceToString(job.Payload()),
			"pop_time", job.PopTime().String(),
		)
	}

	// step2、检查最大尝试执行次数是否超限
	if job.Attempts() >= job.Payload().MaxTries {
		// 超过最大重试次数：本次执行失败 && 任务类最终执行失败 && delete任务
		m.failJob(job, err)
	} else {
		// 任务可以重试：本次执行失败 && 任务类还可以重试 && release任务
		_ = job.Release(job.Payload().RetryInterval)
	}
}

// failJob 失败的任务触发器
func (m *manager) failJob(job JobIFace, err error) {
	// -> 1、标记任务失败
	job.MarkAsFailed()

	// -> 2、任务状态未删除则删除任务
	if job.IsDeleted() {
		return
	}
	_ = job.Delete()

	// tag log
	m.logger.Error(
		textJobFailedLog,
		"queue", job.GetName(),
		"payload", IFaceToString(job.Payload()),
		"error", err.Error(),
	)

	// -> 3、设置任务执行失败
	job.Failed(err)

	// -> 4、queue级别依赖是否有设置失败任务处理器动作
	m.recordFailedJob(job, err)
}

// recordFailedJob 触发记录可能的失败任务
func (m *manager) recordFailedJob(job JobIFace, err error) {
	if m.failedJobHandler != nil {
		_ = m.failedJobHandler(job.Payload(), err)
	}
}

// shutDown 优雅停止队列
// 1、停止轮询loop进程，不再投递job
// 2、上下文设置的等待超时时间内尽量允许执行中的job顺利完成，超时终止的 :reserved 有序队列将在下次执行时再次投递尝试执行
// @param ctx 超时上下文
func (m *manager) shutDown(ctx context.Context) (err error) {
	m.inShutdown.setTrue()

	// 关闭用于控制looper协程的`关闭chan`：这样looper就停止循环
	m.closeDoneChanLocked()

	// 优雅关闭等待时长逐步递增实现
	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}

	m.logger.Info("try graceful shutdown queue, please wait seconds")

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if m.isLooperAndWorkersDown() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}

// getDoneChan 带初始化的获取关闭控制chan
func (m *manager) getDoneChan() <-chan struct{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.getDoneChanLocked()
}

// getDoneChanLocked 底层自动判断的初始化关闭控制chan
func (m *manager) getDoneChanLocked() chan struct{} {
	if m.doneChan == nil {
		m.doneChan = make(chan struct{})
	}
	return m.doneChan
}

// closeDoneChanLocked 关闭用于关闭控制的chan（继而发信号告诉looper和worker优雅停止）
func (m *manager) closeDoneChanLocked() {
	ch := m.getDoneChanLocked()
	select {
	case <-ch:
	default:
		close(ch)
	}
}

// setWorkerStatus 设置标记工作进程当前执行中 or 执行完毕
func (m *manager) setWorkerStatus(workerID int64, isRun bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	node, exist := m.workerStatus[workerID]
	if !exist {
		node = new(atomicBool)
		m.workerStatus[workerID] = node
	}

	if isRun {
		node.setTrue()
	} else {
		node.setFalse()
	}
}

// isLooperAndWorkersDown 检查是否所有worker当前工作任务均处于down状态
func (m *manager) isLooperAndWorkersDown() (down bool) {
	// 所有worker退出
	for _, node := range m.workerStatus {
		if node.isSet() {
			return false
		}
	}

	// looper退出
	return m.isChannelClosed.isSet()
}

// shuttingDown 检测当前队列是否处于正在关闭中的状态
func (m *manager) shuttingDown() bool {
	return m.inShutdown.isSet()
}

// closeChannel 多looper争抢关闭channel
func (m *manager) closeChannel() {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 已标记，不再执行close
	if m.isChannelClosed.isSet() {
		return
	}

	// 标记已关闭
	m.isChannelClosed.setTrue()
	close(m.channel)
}

// isConsumerProcess 判断当前进程实例是否为消费者进程
func (m *manager) isConsumerProcess() bool {
	return m.realTasksNum > 0
}

// autoScaleWorkers 自动检测并扩缩容worker进程
func (m *manager) autoScaleWorkers() error {
	if !m.isConsumerProcess() {
		return fmt.Errorf("queue manager has no workers, maybe this instance is not a consumer process")
	}

	var (
		memSta = m.getMemoryStatistics()
		jobSta = m.getJobStatistics()
	)

	var (
		shouldIncrease  = false
		shouldDecrease  = false
		decreaseNumber  = 0
		increaseNumber  = 0
		maxWorkerNumber = int(int64(m.config.MaxConcurrency)*m.realTasksNum) + 1
		minWorkerNumber = int(m.realTasksNum) + 1
		oneWorkerMemory = memSta.GoMemoryTotal / uint64(m.realTasksNum)
	)

	// ++++++++++++++++++++++++++++++++++++++++++++++++++
	// 1. 待执行job数小于启动的Worker数
	// 2. 启动的Worker数大于真实允许执行的task数（说明扩容过）
	// ++++++++++++++++++++++++++++++++++++++++++++++++++
	if int(jobSta.TotalJobs) < len(m.workerStatus) && len(m.workerStatus) > minWorkerNumber {
		shouldDecrease = true
		decreaseNumber = len(m.workerStatus) - minWorkerNumber
	}

	// ++++++++++++++++++++++++++++++++++++++++++++++++++
	// 1. 待执行的job数大于等于100倍的worker数
	// 2. Worker数没超过可允许的最大并发数
	// 3. 按系统可用内存和go已申请内存与已使用内存计算扩容数
	// ++++++++++++++++++++++++++++++++++++++++++++++++++
	if jobSta.TotalJobs >= m.config.AutoScaleJobThreshold && len(m.workerStatus) < maxWorkerNumber {
		shouldIncrease = true
		// 初步设定扩容一倍Worker 与 最大worker和当前worker差值的较小值
		increaseNumber = min(int(m.realTasksNum), maxWorkerNumber-len(m.workerStatus))
		// 按系统可用内存计算最大可扩容worker数，取最小可扩容数
		increaseNumber = min(int(memSta.SysMemoryAvailable/oneWorkerMemory), increaseNumber)
	}

	// ++++++++++++++++++++++++++++++++++++++++++++++++++
	// 1. 系统可用内存低于当前go已分配的内存时没办法扩容
	// 2. 系统内存使用率大于0.9
	// ++++++++++++++++++++++++++++++++++++++++++++++++++
	if memSta.SysMemoryAvailable < oneWorkerMemory || memSta.SysMemoryUsedPercent >= memoryMaxPercentThreshold {
		m.logger.Warn("autoScaleWorkers.stop", "reason", "memory usage is too big")
		shouldIncrease = false
	}

	if shouldDecrease {
		return m.decreaseWorkers(decreaseNumber)
	}

	if shouldIncrease {
		return m.increaseWorkers(increaseNumber)
	}

	return nil
}

// increaseWorkers 增加worker
func (m *manager) increaseWorkers(num int) error {
	if m.shuttingDown() {
		return ErrQueueClosed
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	for range num {
		m.logger.Info("start.worker", "worker_id", IFaceToString(m.nextWorkerID))
		m.startSingleWorker()
	}

	return nil
}

// decreaseWorkers 减少worker
func (m *manager) decreaseWorkers(num int) error {
	if m.shuttingDown() {
		return ErrQueueClosed
	}

	if len(m.workerChannel) <= num {
		return fmt.Errorf("exist worker num %d less then stop worker num %d", len(m.workerChannel), num)
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	for index, stopChan := range m.workerChannel {
		if index > m.nextWorkerID-int64(num) {
			close(stopChan)
			m.logger.Info("stop.worker", "worker_id", IFaceToString(index))
		}
	}
	m.nextWorkerID -= int64(num)

	return nil
}

// getMemoryStatistics 获取内存统计信息
func (m *manager) getMemoryStatistics() MemoryStatistics {
	// 系统内存情况统计
	var (
		memTotal       = uint64(1) // 默认1 防止获取不到时下方报除以0的panic
		memUsed        = uint64(0)
		memAvailable   = uint64(0)
		memUsedPercent = float64(0)
	)
	v, err := mem.VirtualMemory()
	if err != nil {
		m.logger.Warn("get system memory info occur error", "errorMsg", err.Error())
	} else {
		memTotal = v.Total
		memUsed = v.Used
		memAvailable = v.Available
		memUsedPercent = v.UsedPercent
	}

	// go内存情况统计
	goMem := runtime.MemStats{}
	runtime.ReadMemStats(&goMem)

	return MemoryStatistics{
		SysMemoryTotal:       memTotal,
		SysMemoryUsed:        memUsed,
		SysMemoryAvailable:   memAvailable,
		SysMemoryUsedPercent: memUsedPercent,
		GoMemoryTotal:        goMem.Sys,
		GoMemoryAlloc:        goMem.Alloc,
		GoMemoryUsedPercent:  float64(goMem.Sys) / float64(memTotal),
	}
}

// getWorkerStatistics 获取worker统计信息
func (m *manager) getWorkerStatistics() WorkerStatistics {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 统计活跃worker数量
	activeWorkers := int64(0)
	workerState := make(map[int64]bool)
	for workerId, status := range m.workerStatus {
		if status.isSet() {
			activeWorkers++
		}
		workerState[workerId] = status.isSet()
	}

	return WorkerStatistics{
		ActiveWorkers: activeWorkers,
		TotalWorkers:  int64(len(m.workerStatus)),
		WorkerState:   workerState,
	}
}

// getJobStatistics 获取job统计信息
func (m *manager) getJobStatistics() JobStatistics {
	m.lock.Lock()
	defer m.lock.Unlock()

	// 统计允许执行的job待执行情况
	totalJobs := int64(0)
	jobsStatistics := make(map[string]int64)
	for jobName := range m.tasks {
		if m.allowRun(jobName) {
			jobsStatistics[jobName] = m.queue.Size(jobName)
			totalJobs += jobsStatistics[jobName]
		}
	}

	return JobStatistics{
		TotalJobs:      totalJobs,
		JobsStatistics: jobsStatistics,
	}
}

// getStatistics 获取统计信息
func (m *manager) getStatistics() Statistics {
	if !m.isConsumerProcess() {
		m.logger.Warn("queue manager has no workers, maybe this instance is not a consumer process")
	}

	return Statistics{
		StatisticsTime:   time.Now().Unix(),
		MemoryStatistics: m.getMemoryStatistics(),
		WorkerStatistics: m.getWorkerStatistics(),
		JobStatistics:    m.getJobStatistics(),
	}
}
