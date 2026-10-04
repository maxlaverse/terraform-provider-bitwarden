package bwscli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

var (
	resourceNotFoundRegexp = regexp.MustCompile(`(?m)Resource not found\.`)
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
		default:
			return err
		}
	}
	return err
}

func isObjectNotFoundError(err *command.CommandError) bool {
	return resourceNotFoundRegexp.Match([]byte(err.Stderr()))
}
