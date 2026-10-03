package bwcli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

var (
	attachmentNotFoundRegexp = regexp.MustCompile(`^Attachment .* was not found.$`)
)

func newUnmarshallError(err error, args []string) error {
	return fmt.Errorf("unable to parse result of '%s': %w", strings.Join(args, " "), err)
}

func remapError(err error) error {
	var cmdErr *command.CommandError
	if errors.As(err, &cmdErr) {
		switch {
		case isObjectNotFoundError(cmdErr):
			return models.ErrObjectNotFound
		case isAttachmentNotFoundError(cmdErr):
			return models.ErrAttachmentNotFound
		case strings.Contains(cmdErr.Stderr(), "Username or password is incorrect"):
			return errors.New("Username or password is incorrect")
		case strings.Contains(cmdErr.Stderr(), "Invalid master password"):
			return errors.New("Invalid master password")
		}
	}
	return err
}

func isAttachmentNotFoundError(err *command.CommandError) bool {
	return attachmentNotFoundRegexp.Match([]byte(err.Stderr()))
}

func isObjectNotFoundError(err *command.CommandError) bool {
	return err.Stderr() == "Not found."
}
