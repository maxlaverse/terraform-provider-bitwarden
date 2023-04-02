package test

import (
	"bytes"
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
	stdOut      io.Writer
	stdErr      io.Writer
	callback    func(string)
	dummyOutput map[string]string
}

func (c *testCommand) ClearEnv() executor.Command {
	return c
}

func (c *testCommand) WithEnv(envs []string) executor.Command {
	return c
}

func (c *testCommand) WithStdin(dir string) executor.Command {
	return c
}

func (c *testCommand) WithCombinedOutput(out io.Writer) executor.Command {
	c.stdOut = out
	c.stdErr = out
	return c
}

func (c *testCommand) WithOutput(out io.Writer) executor.Command {
	c.stdOut = out
	return c
}

func (c *testCommand) Run() error {
	var combinedOut bytes.Buffer
	err := c.WithCombinedOutput(&combinedOut).(*testCommand).runnerCmd()
	if err != nil {
		return fmt.Errorf("error running '%s': %v, %v", strings.Join(c.args, " "), err, combinedOut.String())
	}
	return nil
}

func (c *testCommand) RunCaptureOutput() ([]byte, error) {
	var combinedOut bytes.Buffer
	var out bytes.Buffer
	err := c.WithCombinedOutput(&combinedOut).WithOutput(&out).(*testCommand).runnerCmd()
	if err != nil {
		return out.Bytes(), fmt.Errorf("error running '%s': %v, %v", strings.Join(c.args, " "), err, combinedOut.String())
	}

	return out.Bytes(), nil
}

func (c *testCommand) runnerCmd() error {
	argsStr := strings.Join(c.args, " ")
	c.callback(argsStr)
	if v, ok := c.dummyOutput[argsStr]; ok {
		c.stdOut.Write([]byte(v))
		return nil
	}
	return fmt.Errorf("unknown test command: '%s', '%s'", c.cmd, c.args)
}
