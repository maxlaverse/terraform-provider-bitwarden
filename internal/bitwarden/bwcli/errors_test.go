//go:build offline

package bwcli

import (
	"fmt"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemapError(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   error
	}{
		{
			name:   "object not found",
			stderr: "Not found.",
			want:   models.ErrObjectNotFound,
		},
		{
			name:   "attachment not found",
			stderr: "Attachment 123 was not found.",
			want:   models.ErrAttachmentNotFound,
		},
		{
			name:   "invalid credentials",
			stderr: "Username or password is incorrect. Try again.",
			want:   models.ErrInvalidCredentials,
		},
		{
			name:   "invalid master password",
			stderr: "Invalid master password.",
			want:   models.ErrWrongMasterPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdErr := command.NewError(fmt.Errorf("test error"), []string{"login", "--raw", "user@example.com"}, tt.stderr)
			got := remapError(cmdErr)
			require.ErrorIs(t, got, tt.want)
			assert.NotContains(t, got.Error(), "user@example.com")
			assert.NotContains(t, got.Error(), tt.stderr)
		})
	}

	t.Run("unrecognized stderr stays on CommandError", func(t *testing.T) {
		cmdErr := command.NewError(fmt.Errorf("test error"), []string{"login", "--raw", "user@example.com"}, "Some other error message")
		got := remapError(cmdErr)
		require.EqualError(t, got, "'test error' while running 'login --raw'")
		assert.ErrorIs(t, got, cmdErr)
	})
}
