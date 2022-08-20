package test

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/executor"
)

type FakeExecutor struct {
	CommandsExecuted []string
	DummyOutput      map[string]string
}

func (e *FakeExecutor) NewCommand(cmd string, args ...string) executor.Command {
	e.CommandsExecuted = append(e.CommandsExecuted, strings.Join(args, " "))
	return &testCommand{
		cmd:         cmd,
		args:        args,
		dummyOutput: e.DummyOutput,
	}
}

func New(dummyOutput map[string]string) executor.Executor {
	return &FakeExecutor{DummyOutput: dummyOutput}
}

type testCommand struct {
	cmd         string
	args        []string
	stdOut      io.Writer
	stdErr      io.Writer
	dummyOutput map[string]string
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
	err := c.WithCombinedOutput(&combinedOut).CmdRun()
	if err != nil {
		return fmt.Errorf("error running '%s': %v, %v", strings.Join(c.args, " "), err, combinedOut.String())
	}
	return nil
}

func (c *testCommand) RunCaptureOutput() ([]byte, error) {
	var combinedOut bytes.Buffer
	var out bytes.Buffer
	err := c.WithCombinedOutput(&combinedOut).WithOutput(&out).CmdRun()
	if err != nil {
		return out.Bytes(), fmt.Errorf("error running '%s': %v, %v", strings.Join(c.args, " "), err, combinedOut.String())
	}

	return out.Bytes(), nil
}

func (c *testCommand) CmdRun() error {
	if v, ok := c.dummyOutput[strings.Join(c.args, " ")]; ok {
		c.stdOut.Write([]byte(v))
		return nil
	}
	return fmt.Errorf("unknown test command: '%s', '%s'", c.cmd, c.args)
}
