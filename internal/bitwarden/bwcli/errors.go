package bwcli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

var (
	attachmentNotFoundRegexp = regexp.MustCompile(`^Attachment .* was not found.$`)
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
