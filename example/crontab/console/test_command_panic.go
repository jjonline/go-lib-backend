package console

type TestCommandPanic struct {
}

func (t TestCommandPanic) Signature() string {
	return "test_command_panic"
}

func (t TestCommandPanic) Desc() string {
	return "测试执行发生panic"
}

func (t TestCommandPanic) Rule() string {
	return "0 */1 * * * *"
}

func (t TestCommandPanic) Execute() error {
	panic("crontab test panic")
}
