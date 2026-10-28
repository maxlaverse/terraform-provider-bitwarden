//go:build offline

package command

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
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

func TestCommandRunOmitsSecretsFromLogsOnError(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		fmt.Println("SECRET_STDOUT")
		fmt.Fprintln(os.Stderr, "SECRET_STDERR")
		os.Exit(1)
		return
	}

	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &logs)

	cmd := New(os.Args[0], "-test.run=TestCommandRunOmitsSecretsFromLogsOnError", "item", "SECRET_ARG")
	cmd.AppendEnv([]string{"GO_WANT_HELPER_PROCESS=1"})

	_, err := cmd.Run(ctx)
	require.Error(t, err)

	logOut := logs.String()
	assert.NotContains(t, logOut, "SECRET_STDOUT")
	assert.NotContains(t, logOut, "SECRET_STDERR")
	assert.NotContains(t, logOut, "SECRET_ARG")
	assert.NotContains(t, err.Error(), "SECRET_ARG")
}

func TestCommandRunOmitsStdoutFromLogsOnSuccess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		fmt.Println("SECRET_STDOUT")
		os.Exit(0)
		return
	}

	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &logs)

	cmd := New(os.Args[0], "-test.run=TestCommandRunOmitsStdoutFromLogsOnSuccess", "item", "SECRET_ARG")
	cmd.AppendEnv([]string{"GO_WANT_HELPER_PROCESS=1"})

	out, err := cmd.Run(ctx)
	require.NoError(t, err)
	assert.Contains(t, string(out), "SECRET_STDOUT")
	assert.NotContains(t, logs.String(), "SECRET_STDOUT")
	assert.NotContains(t, logs.String(), "SECRET_ARG")
}
