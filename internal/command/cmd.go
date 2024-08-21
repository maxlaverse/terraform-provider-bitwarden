package command

import (
	"bytes"
	"context"
	"io"
	"log"
	"os/exec"
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
}

type Command interface {
	AppendEnv(envs []string) Command
	WithStdin(string) Command
	Run(ctx context.Context) ([]byte, error)
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

func (c *command) Run(ctx context.Context) ([]byte, error) {
	log.Printf("[DEBUG] Running command '%v'\n", c.args)
	var stdOut, stdErr bytes.Buffer

	cmd := exec.CommandContext(ctx, c.binary, c.args...)
	cmd.Env = c.env
	cmd.Stdin = c.stdin
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		log.Printf("[ERROR] Command '%v' finished with error: %v\n", c.args, err)
		log.Printf("[ERROR] Stdout: %v\n", stdOut.String())
		log.Printf("[ERROR] Stderr: %v\n", stdErr.String())

		return nil, NewError(err, c.args, stdOut.String(), stdErr.String())
	}
	log.Printf("[DEBUG] Command '%v' finished with success\n", c.args)

	return stdOut.Bytes(), nil
}
