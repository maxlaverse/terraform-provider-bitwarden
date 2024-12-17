package schema_definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSensisitiveFieldsAreMarkedAsSensitive(t *testing.T) {
	sensitiveFields := []string{}

	for k, v := range BaseSchema(DataSource) {
		if v.Sensitive {
			sensitiveFields = append(sensitiveFields, k)
		}
	}

	for k, v := range LoginSchema(DataSource) {
		if v.Sensitive {
			sensitiveFields = append(sensitiveFields, k)
		}
	}

	assert.ElementsMatch(t, []string{"notes", "field", "password", "username", "totp"}, sensitiveFields)
}
