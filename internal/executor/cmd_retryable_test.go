package executor

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testRetryHandler struct {
	called int
}

func (r *testRetryHandler) IsRetryable(err error, attempt int) bool {
	r.called = r.called + 1
	return strings.Contains(err.Error(), "failing on purpose") && attempt < 3
}

func (r *testRetryHandler) Backoff(attempt int) time.Duration {
	return 0
}

func TestCommandRerunOnMatchinError(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		fmt.Println("test: failing on purpose")
		os.Exit(1)
		return
	}
	retryHandler := &testRetryHandler{}
	exe := NewWithRetries(retryHandler)
	cmd := exe.NewCommand(os.Args[0], "-test.run=TestCommandRerunOnMatchinError")
	cmd.AppendEnv([]string{"GO_WANT_HELPER_PROCESS=1"})
	out, err := cmd.Run()

	assert.NotNil(t, err)
	assert.Error(t, err)
	assert.Equal(t, retryHandler.called, 3)
	assert.Equal(t, "test: failing on purpose\n", string(out))
}

func TestCommandFailsOnUnmatchedError(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		fmt.Println("test: failing for other reasons")
		os.Exit(1)
		return
	}
	retryHandler := &testRetryHandler{}
	exe := NewWithRetries(retryHandler)
	cmd := exe.NewCommand(os.Args[0], "-test.run=TestCommandFailsOnUnmatchedError")
	cmd.AppendEnv([]string{"GO_WANT_HELPER_PROCESS=1"})
	out, err := cmd.Run()

	assert.NotNil(t, err)
	assert.Error(t, err)
	assert.Equal(t, retryHandler.called, 1)
	assert.Equal(t, "test: failing for other reasons\n", string(out))
}
