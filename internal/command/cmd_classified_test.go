//go:build offline

package command

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCommand struct {
	err error
}

func (s stubCommand) AppendEnv([]string) Command { return s }
func (s stubCommand) WithStdin(string) Command   { return s }
func (s stubCommand) Run(context.Context) ([]byte, error) {
	return nil, s.err
}

func TestClassifyMapsRunError(t *testing.T) {
	inner := stubCommand{err: fmt.Errorf("raw")}
	cmd := Classify(inner, func(err error) error {
		return fmt.Errorf("classified: %w", err)
	})

	_, err := cmd.AppendEnv([]string{"FOO=bar"}).Run(t.Context())
	require.EqualError(t, err, "classified: raw")
}

func TestClassifyLeavesSuccessAlone(t *testing.T) {
	_, err := Classify(stubCommand{}, func(error) error {
		t.Fatal("classifier should not run on success")
		return nil
	}).Run(t.Context())
	assert.NoError(t, err)
}
