package command

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NewFn func(binary string, args ...string) Command

// New is only meant to be changed during tests.
var New = func(binary string, args ...string) Command {
	return &command{
		args:   args,
		binary: binary,
	}
}

var LookupPath = func(binary string) (string, error) {
	return exec.LookPath(binary)
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
	ctx = tflog.SetField(ctx, "command", c.args)
	tflog.Debug(ctx, "Running command")

	var stdOut, stdErr bytes.Buffer
	cmd := exec.CommandContext(ctx, c.binary, c.args...)
	cmd.Env = c.env
	cmd.Stdin = c.stdin
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		tflog.Error(ctx, "Command finished with error", map[string]interface{}{"error": err})
		tflog.Trace(ctx, "Command outputs", map[string]interface{}{"stdout": stdOut.String(), "stderr": stdErr.String()})
		return nil, NewError(err, c.args, stdOut.String(), stdErr.String())
	}
	tflog.Debug(ctx, "Command finished with success")
	tflog.Trace(ctx, "Command outputs", map[string]interface{}{"stdout": stdOut.String(), "stderr": stdErr.String()})

	return stdOut.Bytes(), nil
}
