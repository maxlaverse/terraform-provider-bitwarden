package bw

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

var (
	ErrObjectNotFound     = errors.New("object not found")
	ErrAttachmentNotFound = errors.New("attachment not found")

	attachmentNotFoundRegexp = regexp.MustCompile(`^Attachment .* was not found.$`)
)

func newUnmarshallError(err error, cmd string, out []byte) error {
	return fmt.Errorf("unable to parse result of '%s', error: '%v', output: '%v'", cmd, err, string(out))
}

func remapError(err error) error {
	v, ok := err.(*command.CommandError)
	if ok {
		switch {
		case isObjectNotFoundError(v):
			return ErrObjectNotFound
		case isAttachmentNotFoundError(v):
			return ErrAttachmentNotFound
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
