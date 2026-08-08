//go:build offline

package schema_definition

import (
	"testing"

	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensisitiveFieldsAreMarkedAsSensitive(t *testing.T) {
	sensitiveFields := []string{}

	for k, v := range LoginDataSourceSchema().Attributes {
		if v.IsSensitive() {
			sensitiveFields = append(sensitiveFields, k)
		}
	}

	assert.ElementsMatch(t, []string{"notes", "field", "password", "username", "totp"}, sensitiveFields)
}

func TestResourceFieldNestedAttributesAreSensitive(t *testing.T) {
	block, ok := LoginResourceSchema().Blocks[AttributeField].(rsschema.ListNestedBlock)
	require.True(t, ok, "field must be a ListNestedBlock")

	for name, attr := range block.NestedObject.Attributes {
		assert.Truef(t, attr.IsSensitive(), "field.%s should be Sensitive (ListNestedBlock cannot mark the list itself)", name)
	}
}
