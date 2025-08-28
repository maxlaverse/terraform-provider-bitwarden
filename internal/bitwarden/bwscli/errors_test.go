package bwscli

import (
	"fmt"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

func TestRemapError(t *testing.T) {
	tests := []struct {
		name      string
		stderr    string
		expected  error
		errorType string // "notFound", "formatted", or "plain"
	}{
		{
			name:      "resource not found",
			stderr:    "Resource not found.",
			expected:  models.ErrObjectNotFound,
			errorType: "notFound",
		},
		{
			name:      "old format not found",
			stderr:    "Not found.",
			expected:  models.ErrObjectNotFound,
			errorType: "notFound",
		},
		{
			name: "404 resource not found with JSON",
			stderr: `Error:
   0: Received error message from server: [404 Not Found] {"message":"Resource not found.","validationErrors":null,"exceptionMessage":null,"exceptionStackTrace":null,"innerExceptionMessage":null,"object":"error"}

Location:
   crates/bws/src/command/secret.rs:125

Backtrace omitted. Run with RUST_BACKTRACE=1 environment variable to display it.`,
			expected:  models.ErrObjectNotFound,
			errorType: "notFound",
		},
		{
			name:      "access token error",
			stderr:    "Error:\n   0: Access token is not in a valid format",
			expected:  fmt.Errorf("Access token is not in a valid format"),
			errorType: "formatted",
		},
		{
			name: "access token format error detailed",
			stderr: `Error:
   0: Access token is not in a valid format: Error decoding base64: Invalid symbol 61, offset 22.
   1: Error decoding base64: Invalid symbol 61, offset 22.
   2: Invalid symbol 61, offset 22.

Location:
    crates/bws/src/main.rs:69

Backtrace omitted. Run with RUST_BACKTRACE=1 environment variable to display it.`,
			expected:  fmt.Errorf("Access token is not in a valid format: Error decoding base64: Invalid symbol 61, offset 22. Error decoding base64: Invalid symbol 61, offset 22. Invalid symbol 61, offset 22."),
			errorType: "formatted",
		},
		{
			name: "cryptography error",
			stderr: `Error:
   0: Cryptography error, The cipher's MAC doesn't match the expected value
   1: The cipher's MAC doesn't match the expected value

Location:
    crates/bws/src/main.rs:106

Backtrace omitted. Run with RUST_BACKTRACE=1 environment variable to display it.`,
			expected:  fmt.Errorf("Cryptography error, The cipher's MAC doesn't match the expected value The cipher's MAC doesn't match the expected value"),
			errorType: "formatted",
		},
		{
			name:      "plain error message",
			stderr:    "Just a plain error message",
			expected:  fmt.Errorf("'test error' while running 'test': , Just a plain error message"),
			errorType: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdErr := command.NewError(fmt.Errorf("test error"), []string{"test"}, "", tt.stderr)

			result := remapError(cmdErr)

			switch tt.errorType {
			case "notFound":
				// For not found errors, check the error type
				if result != models.ErrObjectNotFound {
					t.Errorf("remapError() = %v, want %v", result, tt.expected)
				}
			case "formatted":
				// For formatted errors, check the error message
				if result.Error() != tt.expected.Error() {
					t.Errorf("remapError() = %v, want %v", result.Error(), tt.expected.Error())
				}
			case "plain":
				// For plain errors, check the error message
				if result.Error() != tt.expected.Error() {
					t.Errorf("remapError() = %v, want %v", result.Error(), tt.expected.Error())
				}
			default:
				t.Errorf("unknown error type: %s", tt.errorType)
			}
		})
	}
}
