# Queue 队列

## 一、说明

Queue队列为`生产 -> 消费`模型的简单实现，即：`producer -> consumer(worker)`，一般分为生产端和消费端。

当前已实现开发测试用`memory`方案和可用于生产的`redis`方案。

> **由于多个独立进程间内存隔离，以及进程退出后进程所属内存销毁的原因，`memory`方案在进程退出后未消费的队列数据会丢失，故而仅能用于开发调试环境，且生产端和消费端只能在同一进程。**

## 二、版本差异说明

## v0.7.x系列

`v0.7.x`系列底层使用了 [redis-v7](https://github.com/go-redis/redis/releases/tag/v7.4.0)

go-redis库的v7版本和v8版本存在极其大的差异，为了兼容老项目提供`v0.7.x`系列版本。

**请注意：`v0.7.x`系列任务类不支持超时控制。**

## 非v0.7.x系列

`v0.7.x`版本号之外均使用go-redis v8，并且任务类支持`context.Context`上下文控制和超时控制。

## 三、使用示例

完整使用示例查看 [example](https://github.com/jjonline/go-mod-library/tree/master/queue/example) 目录代码结构

### step1、实现任务类

> 任务类即按任务类`iface`规则实现的结构体，也是队列投递任务和实际执行任务的单元。

````
package tasks

import (
    "fmt"
    "github.com/jjonline/go-mod-librar/queue"
)

// 定义的任务类struct，需完整实现 queue.TaskIFace
type TestTask struct {
    // 单个job最大执行时长、最大重试次数、多次重试之间间隔时长等设置
    // 这里使用默认设置，若需要自定义参数，自定义方法实现即可
    queue.DefaultTaskSetting
}

func (t TestTask) Name() string {
    return "test_task"
}

func (t TestTask) Execute(job *queue.RawBody) error {
    // 队列实际执行的入口方法
    fmt.Println(job.ID)
    return nil
}
````

### step2、消费者端注册启动

````
// 初始化队列Queue对象，生产者、消费者均通过该对象操作
// 重要：生产者、消费者均需要实例化
service := queue.New(
    queue.Redis, // 队列底层驱动器类型，详见包内常量
    redisClient, // 队列底层驱动client实例
    zapLogger, // zap日志实例，用于记录日志
    5, // 单个队列最大并发消费协程数
)

// 注册单个任务类
_ = service.BootstrapOne(&tasks.TestTask{})

// 也可以这样批量注册任务类
// _ = service.Bootstrap([]*queue.TaskIFace)

// 启动消费端进程，注意传递上下文context用于控制进程优雅控制
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

    zapLogger.Info("receive exit signal")

    // shutdown worker daemon with timeout context
    timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // graceful shutdown by signal
    if err := queueService.ShutDown(timeoutCtx); nil != err {
        zapLogger.Warn("violence shutdown by signal: " + err.Error())
    } else {
        zapLogger.Info("graceful shutdown by signal")
    }

    // closer close
    close(idleCloser)
}()

// start worker daemon
if err := queueService.Start(); nil != err && err != queue.ErrQueueClosed {
    zapLogger.Info("queue started failed: " + err.Error())
    close(idleCloser)
} else {
    zapLogger.Info("queue worker started")
}

<-idleCloser
zapLogger.Info("queue worker quit, daemon exited")
````

### step3、生产者端投递job任务

````
// 初始化队列Queue对象，生产者、消费者均通过该对象操作
// 生产者&&消费者处于同一进程则可共用，不同进程则需要各自独立实例化
service := queue.New(
    queue.Redis, // 队列底层驱动器类型，详见包内常量
    redisClient, // 队列底层驱动client实例
    zapLogger, // zap日志实例，用于记录日志
)

// 投递一条普通队列任务
service.Dispatch(&tasks.TestTask{}, "job执行时的参数")

// 投递一条延迟队列任务（指定执行时刻）
// 指定执行时刻，如果时刻是过去则立即执行
service.DelayAt(&tasks.TestTask{}, "job执行时的参数", time.Time类型的延迟到将来时刻)

// 投递一条延迟队列任务（指定相对于当前的延迟时长）
// 指定相对于投递时刻需要延迟的时长
service.Delay(&tasks.TestTask{}, "job执行时的参数", time.Duration类型的时长)
````

## 四、重试次数 & 重试间隔 & 超时

> **队列保证每个job至少能被执行1次**

### 3.1、重试次数

任务类定义实现的 `MaxTries() int64` 方法指定单个job能被重试的次数

**注意：返回值若小于等于1则仅被执行1次**

> 执行任务类失败或异常会触发重试

### 3.2、重试间隔

当任务类允许多次重试时，下一次重试可以并不是立即执行，通过`RetryInterval() int64`方法设置重试之前的等待时长间隔，单位：秒

**注意：返回值若小于等于0则取值0表示立即重试**

> `重试间隔` 是配合 `重试次数` 起作用的，仅可多次重试的任务有效

### 3.3、超时

`v0.7.x`系列版本并不支持超时设置，只有一个900秒执行的提示 

> 因goroutine无法从外部kill掉，超时控制通过`context.Context`上下文实现，需任务类自主实现超时控制的退出机制！

默认任务类设置最大超时时长为`900秒`，可通过任务类Timeout方法自定义超时时间。

### 3.4、约定

1. `重试次数`若小于等于1则取值1
2. `重试间隔`若小于等于0则取值0，0表示没有重试间隔
3. 任务执行成功：`Execute(job *RawBody) error`返回`nil`
4. 任务执行失败：`Execute(job *RawBody) error`返回`error`
5. 任务执行异常：`Execute(job *RawBody) error`发生了`panic`
