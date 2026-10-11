package command

import (
	"errors"
	"fmt"
	"strings"
)

type CommandError struct {
	err    error
	args   []string
	stderr string
}

func NewError(err error, args []string, stderr string) *CommandError {
	return &CommandError{
		err:    err,
		args:   args,
		stderr: stderr,
	}
}

func (c *CommandError) Error() string {
	// Do not include stdout, stderr, or argv beyond the subcommand: CLI
	// payloads and write arguments can contain decrypted secrets.
	return fmt.Sprintf("'%s' while running '%s'", c.err, commandSummary(c.args))
}

func (c *CommandError) Stderr() string {
	return c.stderr
}

func (c *CommandError) Unwrap() error {
	return c.err
}

// StderrOrError returns CLI stderr when err is a CommandError, otherwise err.Error().
// Use this for classification (not-found, rate limits), not for diagnostics.
func StderrOrError(err error) string {
	var cmdErr *CommandError
	if errors.As(err, &cmdErr) && cmdErr.stderr != "" {
		return cmdErr.stderr
	}
	return err.Error()
}

func commandSummary(args []string) string {
	if len(args) <= 2 {
		return strings.Join(args, " ")
	}
	return strings.Join(args[:2], " ")
}
