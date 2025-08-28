//go:build offline

package bwscli

import (
	"fmt"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

func TestRemapError(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected error
	}{
		{
			name: "resource not found",
			stderr: `Error:
           0: Received error message from server: [404 Not Found] {"exceptionMessage":null,"exceptionStackTrace":null,"innerExceptionMessage":null,"message":"Resource not found. (project): 283e673a-2b95-46b7-9a1e-89b4fcf56f24","object":"error","validationErrors":null}`,
			expected: models.ErrObjectNotFound,
		},
		{
			name:     "anything but not found",
			stderr:   "Some other error message",
			expected: fmt.Errorf("'test error' while running 'test': , Some other error message"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdErr := command.NewError(fmt.Errorf("test error"), []string{"test"}, "", tt.stderr)

			result := remapError(cmdErr)

			if result.Error() != tt.expected.Error() {
				t.Errorf("remapError() = %v, want %v", result, tt.expected)
			}
		})
	}
}
