package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

func New(dummyOutput map[string]string, callback func(string, *string)) command.NewFn {
	return func(cmd string, args ...string) command.Command {
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
	stdin       *string
	callback    func(string, *string)
	dummyOutput map[string]string
}

func (c *testCommand) AppendEnv(envs []string) command.Command {
	return c
}

func (c *testCommand) WithStdin(data string) command.Command {
	c.stdin = &data
	return c
}

func (c *testCommand) Run(_ context.Context) ([]byte, error) {
	argsStr := strings.Join(c.args, " ")
	c.callback(argsStr, c.stdin)

	if v, ok := c.dummyOutput[argsStr]; ok {
		return []byte(v), nil
	}
	if v, ok := c.dummyOutput[strings.Join(append(c.args, "@error"), " ")]; ok {
		return nil, fmt.Errorf("failing command '%s' for test purposes: %v", argsStr, v)
	}
	return nil, fmt.Errorf("[unknown test command: '%s', '%s'", c.cmd, c.args)
}

func MockCommands(t *testing.T, dummyOutput map[string]string) (func(t *testing.T), func() []string) {
	commandsExecuted := []string{}
	newCommandToRestore := command.New
	command.New = New(dummyOutput, func(args string, stdin *string) {
		if stdin != nil {
			commandsExecuted = append(commandsExecuted, fmt.Sprintf("%s:/:%s", *stdin, args))
		} else {
			commandsExecuted = append(commandsExecuted, args)
		}
	})
	return func(t *testing.T) {
			command.New = newCommandToRestore
		},
		func() []string {
			return commandsExecuted
		}
}
