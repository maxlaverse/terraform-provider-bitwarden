package command

import "context"

// Classify wraps cmd so Run() errors are passed through fn.
func Classify(cmd Command, fn func(error) error) Command {
	return classifiedCommand{Command: cmd, fn: fn}
}

type classifiedCommand struct {
	Command
	fn func(error) error
}

func (c classifiedCommand) AppendEnv(envs []string) Command {
	c.Command = c.Command.AppendEnv(envs)
	return c
}

func (c classifiedCommand) WithStdin(stdin string) Command {
	c.Command = c.Command.WithStdin(stdin)
	return c
}

func (c classifiedCommand) Run(ctx context.Context) ([]byte, error) {
	out, err := c.Command.Run(ctx)
	if err != nil && c.fn != nil {
		return out, c.fn(err)
	}
	return out, err
}
