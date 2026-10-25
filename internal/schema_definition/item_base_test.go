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

func TestSecretValueAndNoteAreSensitive(t *testing.T) {
	for _, name := range []string{AttributeValue, AttributeNote} {
		attr, ok := SecretResourceSchema().Attributes[name]
		require.Truef(t, ok, "resource %s", name)
		assert.Truef(t, attr.IsSensitive(), "resource %s should be Sensitive", name)

		attr, ok = SecretDataSourceSchema().Attributes[name]
		require.Truef(t, ok, "data source %s", name)
		assert.Truef(t, attr.IsSensitive(), "data source %s should be Sensitive", name)
	}
}

func TestAttachmentContentAndURLAreSensitive(t *testing.T) {
	content, ok := AttachmentResourceSchema().Attributes[AttributeAttachmentContent]
	require.True(t, ok)
	assert.True(t, content.IsSensitive(), "resource content should be Sensitive")

	url, ok := AttachmentResourceSchema().Attributes[AttributeAttachmentURL]
	require.True(t, ok)
	assert.True(t, url.IsSensitive(), "resource url should be Sensitive")

	dsContent, ok := AttachmentDataSourceSchema().Attributes[AttributeAttachmentContent]
	require.True(t, ok)
	assert.True(t, dsContent.IsSensitive(), "data source content should be Sensitive")
}

func TestSecretKeyDescriptionIsNotGenericName(t *testing.T) {
	attr, ok := SecretResourceSchema().Attributes[AttributeKey]
	require.True(t, ok)
	assert.Contains(t, attr.GetMarkdownDescription(), "Secrets Manager secret")
	assert.NotEqual(t, "Name.", attr.GetMarkdownDescription())
}

func TestLoginTotpDescriptionIsTheSeed(t *testing.T) {
	attr, ok := LoginResourceSchema().Attributes[AttributeLoginTotp]
	require.True(t, ok)
	assert.Contains(t, attr.GetDescription(), "seed")
	assert.NotEqual(t, "Verification code.", attr.GetDescription())
}
