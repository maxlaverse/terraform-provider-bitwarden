package bwscli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

var (
	attachmentNotFoundRegexp = regexp.MustCompile(`^Attachment .* was not found.$`)
	resourceNotFoundRegexp   = regexp.MustCompile(`Resource not found\.`)
)

func newUnmarshallError(err error, args []string, out []byte) error {
	return fmt.Errorf("unable to parse result of '%s', error: '%v', output: '%v'", strings.Join(args, " "), err, string(out))
}

func remapError(err error) error {
	v, ok := err.(*command.CommandError)
	if ok {
		switch {
		case isObjectNotFoundError(v):
			return models.ErrObjectNotFound
		case isAttachmentNotFoundError(v):
			return models.ErrAttachmentNotFound
		default:
			// For other errors, reformat them to show just the error message
			return reformatBwsError(err)
		}
	}
	return err
}

func isAttachmentNotFoundError(err *command.CommandError) bool {
	return attachmentNotFoundRegexp.Match([]byte(err.Stderr()))
}

func isObjectNotFoundError(err *command.CommandError) bool {
	// Check for both old "Not found." and new "Resource not found." errors
	return err.Stderr() == "Not found." || resourceNotFoundRegexp.Match([]byte(err.Stderr()))
}

// reformatBwsError extracts just the error message from BWS CLI output,
// removing the "Location:" and backtrace information
func reformatBwsError(err error) error {
	if cmdErr, ok := err.(*command.CommandError); ok {
		stderr := cmdErr.Stderr()

		// Look for the error message after "Error:" and before "Location:"
		if strings.Contains(stderr, "Error:") {
			parts := strings.Split(stderr, "Location:")
			if len(parts) > 0 {
				errorPart := strings.TrimSpace(parts[0])
				// Remove the "Error:" prefix and clean up
				if strings.HasPrefix(errorPart, "Error:") {
					errorPart = strings.TrimSpace(strings.TrimPrefix(errorPart, "Error:"))
				}

				// Remove numbered prefixes like "0:", "1:", "2:" and clean up whitespace
				lines := strings.Split(errorPart, "\n")
				var cleanedLines []string
				for _, line := range lines {
					line = strings.TrimSpace(line)
					// Remove numbered prefixes (0:, 1:, 2:, etc.)
					if strings.Contains(line, ":") {
						colonIndex := strings.Index(line, ":")
						if colonIndex > 0 {
							prefix := strings.TrimSpace(line[:colonIndex])
							// Check if prefix is just a number followed by colon
							if len(prefix) <= 3 && strings.TrimSuffix(prefix, ":") == strings.TrimSuffix(prefix, ":") {
								line = strings.TrimSpace(line[colonIndex+1:])
							}
						}
					}
					if line != "" {
						cleanedLines = append(cleanedLines, line)
					}
				}

				errorPart = strings.Join(cleanedLines, " ")
				return fmt.Errorf("%s", errorPart)
			}
		}
	}
	return err
}
