package executor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type Executor interface {
	NewCommand(cmd string, args ...string) Command
}

func New() Executor {
	return &executor{}
}

type executor struct{}

func (e *executor) NewCommand(cmd string, args ...string) Command {
	return &command{
		cmd: exec.Command(cmd, args...),
	}
}

type command struct {
	cmd *exec.Cmd
}

type Command interface {
	WithEnv(envs []string) Command
	WithOutput(out io.Writer) Command
	WithCombinedOutput(out io.Writer) Command
	WithStdin(string) Command
	Run() error
	RunCaptureOutput() ([]byte, error)
	cmdRun() error
}

func (c *command) WithEnv(envs []string) Command {
	c.cmd.Env = append(c.cmd.Env, envs...)
	return c
}
func (c *command) WithStdin(dir string) Command {
	var stdinBuf bytes.Buffer
	stdinBuf.Write([]byte(dir))

	c.cmd.Stdin = &stdinBuf
	return c
}

func (c *command) WithCombinedOutput(out io.Writer) Command {
	if c.cmd.Stdout != nil {
		c.cmd.Stdout = io.MultiWriter(c.cmd.Stdout, out)
	} else {
		c.cmd.Stdout = out
	}
	if c.cmd.Stderr != nil {
		c.cmd.Stderr = io.MultiWriter(c.cmd.Stderr, out)
	} else {
		c.cmd.Stderr = out
	}
	return c
}

func (c *command) WithOutput(out io.Writer) Command {
	if c.cmd.Stdout != nil {
		c.cmd.Stdout = io.MultiWriter(c.cmd.Stdout, out)
	} else {
		c.cmd.Stdout = out
	}
	return c
}

func (c *command) Run() error {
	var combinedOut bytes.Buffer
	err := c.WithCombinedOutput(&combinedOut).cmdRun()
	if err != nil {
		return fmt.Errorf("error running '%s': %v, %v", strings.Join(c.cmd.Args, " "), err, combinedOut.String())
	}
	return nil
}

func (c *command) RunCaptureOutput() ([]byte, error) {
	var combinedOut bytes.Buffer
	var out bytes.Buffer
	err := c.WithCombinedOutput(&combinedOut).WithOutput(&out).cmdRun()
	if err != nil {
		return out.Bytes(), fmt.Errorf("error running '%s': %v, %v", strings.Join(c.cmd.Args, " "), err, combinedOut.String())
	}

	return out.Bytes(), nil
}

func (c *command) cmdRun() error {
	log.Printf("Running %v\n", c.cmd.Args)
	return c.cmd.Run()
}
