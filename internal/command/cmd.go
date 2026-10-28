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
	// Never attach argv, stdout, or stderr to logs: bw/bws payloads include
	// decrypted secrets. commandSummary is the subcommand only (e.g. "create item").
	fields := map[string]interface{}{
		"binary":  c.binary,
		"command": commandSummary(c.args),
	}
	tflog.Debug(ctx, "Running command", fields)

	var stdOut, stdErr bytes.Buffer
	cmd := exec.CommandContext(ctx, c.binary, c.args...)
	cmd.Env = c.env
	cmd.Stdin = c.stdin
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		tflog.Error(ctx, "Command finished with error", map[string]interface{}{"error": err, "binary": c.binary, "command": commandSummary(c.args)})
		return nil, NewError(err, c.args, stdErr.String())
	}
	tflog.Debug(ctx, "Command finished with success", fields)

	return stdOut.Bytes(), nil
}
