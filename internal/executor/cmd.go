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
	RunSingle() error
	RunCaptureCombined() ([]byte, error)
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
	c.cmd.Stdout = out
	c.cmd.Stderr = out
	return c
}

func (c *command) WithOutput(out io.Writer) Command {
	c.cmd.Stdout = out
	return c
}

func (c *command) RunSingle() error {
	log.Printf("Running %v\n", c.cmd.Args)
	return c.cmd.Run()
}

func (c *command) Run() error {
	var out bytes.Buffer
	err := c.WithCombinedOutput(&out).RunSingle()
	if err != nil {
		return fmt.Errorf("error running '%s': %v, %v", strings.Join(c.cmd.Args, " "), err, out.String())
	}

	return nil
}

func (c *command) RunCaptureCombined() ([]byte, error) {
	var out bytes.Buffer
	err := c.WithCombinedOutput(&out).RunSingle()
	if err != nil {
		return out.Bytes(), fmt.Errorf("error running '%s': %v, %v", strings.Join(c.cmd.Args, " "), err, out.String())
	}

	return out.Bytes(), nil

}
