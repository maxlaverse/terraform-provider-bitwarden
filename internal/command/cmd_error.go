package command

import (
	"fmt"
	"strings"
)

type CommandError struct {
	err    error
	args   []string
	stdout string
	stderr string
}

func NewError(err error, args []string, stdout, stderr string) *CommandError {
	return &CommandError{
		err:    err,
		args:   args,
		stdout: stdout,
		stderr: stderr,
	}
}

func (c CommandError) Error() string {
	return fmt.Sprintf("'%s' while running '%s': %v, %v", c.err, strings.Join(c.args, " "), c.stdout, c.stderr)
}

func (c CommandError) Stderr() string {
	return c.stderr
}
