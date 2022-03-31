package console

type TestCommandFail struct {
}

func (t TestCommandFail) Signature() string {
	return "test_command_fail"
}

func (t TestCommandFail) Rule() string {
	return "0 */1 * * * *"
}

func (t TestCommandFail) Execute() error {
	return nil
}
