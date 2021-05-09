package console

type TestCommandFail struct {

}

func (t TestCommandFail) Signature() string {
	return "test_command_fail"
}

func (t TestCommandFail) Description() string {
	return "this is a test crontab command implement"
}

func (t TestCommandFail) Execute() error {
	return nil
}

