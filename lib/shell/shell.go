package shell

type Shell interface {
	NewCommand(cmd string) Command
}

type ShellImpl struct {
}

func (s ShellImpl) NewCommand(cmd string) Command {
	return &command{
		program:    DefaultProgram,
		outputType: DefaultOutputType,
		cmd:        cmd,
		timeout:    DefaultTimeout,
	}
}
