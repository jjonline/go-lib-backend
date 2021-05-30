package contract

// Command 定时任务&&cli命令行契约
type Command interface {
	Signature() string              // Signature   命令行签名参数
	Description() string            // Description 返回命令行描述
	Execute(args ...[]string) error // Execute     命令执行入口
}
