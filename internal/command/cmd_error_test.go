//go:build offline

package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandErrorOmitsSecrets(t *testing.T) {
	err := NewError(fmt.Errorf("exit status 1"), []string{"secret", "create", "KEY", "THESECRET", "proj"}, "Resource not found.")

	assert.Equal(t, "'exit status 1' while running 'secret create'", err.Error())
	assert.NotContains(t, err.Error(), "THESECRET")
	assert.Equal(t, "Resource not found.", err.Stderr())
}

func TestCommandErrorWithoutStderr(t *testing.T) {
	err := NewError(fmt.Errorf("exit status 1"), []string{"status"}, "")

	assert.Equal(t, "'exit status 1' while running 'status'", err.Error())
	assert.Empty(t, err.Stderr())
}

func TestCommandErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("exit status 1")
	err := NewError(cause, []string{"status"}, "Rate limit exceeded.")

	require.ErrorIs(t, err, cause)
}

func TestStderrOrError(t *testing.T) {
	cmdErr := NewError(fmt.Errorf("exit status 1"), []string{"status"}, "Rate limit exceeded.")
	assert.Equal(t, "Rate limit exceeded.", StderrOrError(cmdErr))
	assert.Equal(t, "plain", StderrOrError(fmt.Errorf("plain")))
}
