package command

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type NewFn func(binary string, args ...string) Command

// New is only meant to be changed during tests.
var New = func(binary string, args ...string) Command {
	return &command{
		args:   args,
		binary: binary,
	}
}

type command struct {
	binary string
	args   []string
	env    []string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type Command interface {
	AppendEnv(envs []string) Command
	WithOutput(out io.Writer) Command
	WithStdin(string) Command
	Run() ([]byte, error)
}

func (c *command) AppendEnv(envs []string) Command {
	c.env = append(c.env, envs...)
	return c
}
func (c *command) WithStdin(dir string) Command {
	var stdinBuf bytes.Buffer
	stdinBuf.Write([]byte(dir))

	c.stdin = &stdinBuf
	return c
}

func (c *command) WithOutput(out io.Writer) Command {
	c.stdout = out
	return c
}

func (c *command) Run() ([]byte, error) {
	log.Printf("[DEBUG] Running command '%v'\n", c.args)
	cmd := exec.Command(c.binary, c.args...)
	cmd.Stdin = c.stdin
	cmd.Env = c.env

	var combinedOut bytes.Buffer
	var out bytes.Buffer
	if c.stdout != nil {
		cmd.Stdout = io.MultiWriter(c.stdout, &combinedOut, &out)
	} else {
		cmd.Stdout = io.MultiWriter(&combinedOut, &out)
	}

	if c.stderr != nil {
		cmd.Stderr = io.MultiWriter(c.stderr, &combinedOut)
	} else {
		cmd.Stderr = io.MultiWriter(&combinedOut)
	}

	err := cmd.Run()
	if err != nil {
		log.Printf("[ERROR] Command '%v' finished with error: %v\n", c.args, err)
		return out.Bytes(), fmt.Errorf("error running '%s': %v, %v", strings.Join(c.args, " "), err, combinedOut.String())
	}
	log.Printf("[DEBUG] Command '%v' finished with success\n", c.args)

	return out.Bytes(), nil
}
