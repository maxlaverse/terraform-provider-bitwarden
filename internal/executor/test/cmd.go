package test

import (
	"fmt"
	"io"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/executor"
)

func New(dummyOutput map[string]string, callback func(string)) executor.NewCommandFn {
	return func(cmd string, args ...string) executor.Command {
		return &testCommand{
			cmd:         cmd,
			args:        args,
			callback:    callback,
			dummyOutput: dummyOutput,
		}
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
	if v, ok := c.dummyOutput[strings.Join(append(c.args, "@error"), " ")]; ok {
		return nil, fmt.Errorf("failing command '%s' for test purposes: %v", argsStr, v)
	}
	return nil, fmt.Errorf("[unknown test command: '%s', '%s'", c.cmd, c.args)
}
