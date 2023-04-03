package command

import (
	"log"
	"time"
)

type RetryHandler interface {
	IsRetryable(err error, attempts int) bool
	Backoff(attempt int) time.Duration
}

func NewWithRetries(retryHandler RetryHandler) NewFn {
	return func(binary string, args ...string) Command {
		return &retryableCommand{
			cmd:          New(binary, args...),
			retryHandler: retryHandler,
		}
	}
}

type retryableCommand struct {
	cmd          Command
	retryHandler RetryHandler
}

func (c *retryableCommand) AppendEnv(envs []string) Command {
	c.cmd.AppendEnv(envs)
	return c
}
func (c *retryableCommand) WithStdin(dir string) Command {
	c.cmd.WithStdin(dir)
	return c
}

func (c *retryableCommand) Run() ([]byte, error) {
	attempts := 0
	for {
		attempts = attempts + 1
		out, err := c.cmd.Run()
		if err == nil || !c.retryHandler.IsRetryable(err, attempts) {
			return out, err
		}
		c.retryHandler.Backoff(attempts)
		log.Printf("[ERROR] Retrying command after error: %v\n", err)
	}
}
