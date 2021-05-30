package console

type TestCommandOk struct {

}

func (t TestCommandOk) Signature() string {
	return "test_command_ok"
}

func (t TestCommandOk) Description() string {
	return "this is a test crontab command implement"
}

func (t TestCommandOk) Execute(args ...[]string) error {
	return nil
}

