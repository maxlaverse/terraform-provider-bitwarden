//go:build offline

package command

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandRunOmitsOutputFromError(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		fmt.Println("private-stdout")
		fmt.Fprintln(os.Stderr, "visible-stderr")
		os.Exit(1)
		return
	}

	cmd := New(os.Args[0], "-test.run=TestCommandRunOmitsOutputFromError")
	cmd.AppendEnv([]string{"GO_WANT_HELPER_PROCESS=1"})

	_, err := cmd.Run(t.Context())
	require.Error(t, err)

	assert.NotContains(t, err.Error(), "private-stdout")
	assert.NotContains(t, err.Error(), "visible-stderr")
	assert.Contains(t, err.(*CommandError).Stderr(), "visible-stderr")
	assert.NotContains(t, err.(*CommandError).Stderr(), "private-stdout")
}
