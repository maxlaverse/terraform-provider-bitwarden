//go:build offline

package bwcli

import (
	"fmt"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
	"github.com/stretchr/testify/assert"
)

func TestRemapError(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected string
	}{
		{
			name:     "object not found",
			stderr:   "Not found.",
			expected: models.ErrObjectNotFound.Error(),
		},
		{
			name:     "attachment not found",
			stderr:   "Attachment 123 was not found.",
			expected: models.ErrAttachmentNotFound.Error(),
		},
		{
			name:     "invalid username or password",
			stderr:   "Username or password is incorrect. Try again.",
			expected: "Username or password is incorrect",
		},
		{
			name:     "invalid master password",
			stderr:   "Invalid master password.",
			expected: "Invalid master password",
		},
		{
			name:     "unrecognized stderr stays on CommandError",
			stderr:   "Some other error message",
			expected: "'test error' while running 'login --raw'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdErr := command.NewError(fmt.Errorf("test error"), []string{"login", "--raw", "user@example.com"}, tt.stderr)
			got := remapError(cmdErr)
			assert.Equal(t, tt.expected, got.Error())
			assert.NotContains(t, got.Error(), "user@example.com")
			assert.NotContains(t, got.Error(), tt.stderr)
		})
	}
}
