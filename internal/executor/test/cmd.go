package test

import (
	"fmt"
	"io"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/executor"
)

type FakeExecutor struct {
	Callback    func(string)
	DummyOutput map[string]string
}

func (e *FakeExecutor) NewCommand(cmd string, args ...string) executor.Command {
	return &testCommand{
		cmd:         cmd,
		args:        args,
		callback:    e.Callback,
		dummyOutput: e.DummyOutput,
	}
}

func New(dummyOutput map[string]string, callback func(string)) executor.Executor {
	return &FakeExecutor{
		DummyOutput: dummyOutput,
		Callback:    callback,
	}
}

type testCommand struct {
	cmd         string
	args        []string
	callback    func(string)
	dummyOutput map[string]string
}

func (c *testCommand) AppendEnv(envs []string) executor.Command {
	return c
}

func (c *testCommand) WithStdin(dir string) executor.Command {
	return c
}

func (c *testCommand) WithOutput(out io.Writer) executor.Command {
	return c
}

func (c *testCommand) Run() ([]byte, error) {
	argsStr := strings.Join(c.args, " ")
	c.callback(argsStr)

	if v, ok := c.dummyOutput[argsStr]; ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("unknown test command: '%s', '%s'", c.cmd, c.args)
}
