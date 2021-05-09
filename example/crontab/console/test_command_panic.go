package console

type TestCommandPanic struct {

}

func (t TestCommandPanic) Signature() string {
	return "test_command_panic"
}

func (t TestCommandPanic) Description() string {
	return "this is a test crontab command implement"
}

func (t TestCommandPanic) Execute() error {
	panic("crontab test panic")
}

