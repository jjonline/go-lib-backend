package console

type TestCommandOk struct {
}

func (t TestCommandOk) Signature() string {
	return "test_command_ok"
}

func (t TestCommandOk) Rule() string {
	return "0 */1 * * * *"
}

func (t TestCommandOk) Execute() error {
	return nil
}
