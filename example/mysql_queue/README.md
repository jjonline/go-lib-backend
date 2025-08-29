# MySQL队列驱动示例

## 说明

这个示例展示了如何使用MySQL作为队列驱动的完整用法。

## 准备工作

### 1. 数据库准备

首先需要创建MySQL数据库并执行建表语句：

```sql
-- 创建数据库
CREATE DATABASE test_queue DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

-- 使用数据库
USE test_queue;

-- 创建队列表
-- 执行 ../../queue/stubs/mysql_queue_tables.sql 中的建表语句
```

### 2. 修改数据库连接

在 `main.go` 中修改数据库连接字符串：

```go
dsn := "root:password@tcp(localhost:3306)/test_queue?charset=utf8mb4&parseTime=True&loc=Local"
```

将 `root:password` 替换为你的MySQL用户名和密码。

## 运行示例

```bash
# 进入示例目录
cd example/mysql_queue

# 安装依赖
go mod tidy

# 运行示例
go run main.go
```

## 功能特性

### 支持的队列操作

1. **普通队列任务** - 立即执行的任务
2. **延迟队列任务** - 指定时间后执行的任务
3. **任务重试** - 失败任务的自动重试机制
4. **并发处理** - 支持多协程并发消费任务
5. **优雅关闭** - 支持优雅关闭队列服务

### MySQL队列实现原理

- 使用 `queue_jobs` 表存储所有任务
- `available_at` 字段控制任务的可执行时间（延迟队列支持）
- `reserved_at` 字段标记任务被消费者获取的时间
- `attempts` 字段记录任务尝试次数
- 失败任务可选择记录到 `queue_failed_jobs` 表

### 队列状态流转

1. **投递阶段**：任务插入到 `queue_jobs` 表，`reserved_at` 为 NULL
2. **获取阶段**：消费者获取任务时设置 `reserved_at` 为超时时间戳
3. **执行阶段**：任务成功执行后从表中删除
4. **重试阶段**：任务失败时根据策略重新设置 `available_at`
5. **失败阶段**：超过重试次数的任务记录到失败表

## 注意事项

1. 确保MySQL数据库版本支持 `FOR UPDATE` 语句
2. 建议在 `queue_name`、`available_at`、`reserved_at` 字段上创建索引以提高性能
3. 定期清理 `queue_failed_jobs` 表中的历史失败记录
4. 在高并发场景下，可考虑使用数据库连接池优化性能
