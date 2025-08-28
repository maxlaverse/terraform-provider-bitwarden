package bwscli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

var (
	resourceNotFoundRegexp = regexp.MustCompile(`(?m)Resource not found\.`)
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
		default:
			return err
		}
	}
	return err
}

func isObjectNotFoundError(err *command.CommandError) bool {
	return resourceNotFoundRegexp.Match([]byte(err.Stderr()))
}
