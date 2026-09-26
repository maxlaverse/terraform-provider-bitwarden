//go:build offline

package schema_definition

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashedFileSemanticEquals_LegacySHA1MatchesPath(t *testing.T) {
	path := filepath.Join("..", "provider", "fixtures", "attachment1.txt")
	digest, err := FileSha1Sum(path)
	require.NoError(t, err)
	assert.Equal(t, "34945801b5aed4540ccfde8320ec7c395325e02d", digest)

	legacy := NewHashedFileValue(digest)
	configured := NewHashedFileValue(path)

	equal, diags := legacy.StringSemanticEquals(t.Context(), configured)
	require.False(t, diags.HasError(), diags.Errors())
	assert.True(t, equal, "legacy SDKv2 SHA1 state must equal matching file path (no recreate on upgrade)")
}

func TestHashedFileSemanticEquals_SameContentDifferentPath(t *testing.T) {
	a := NewHashedFileValue(filepath.Join("..", "provider", "fixtures", "attachment1.txt"))
	b := NewHashedFileValue(filepath.Join("..", "provider", "fixtures", "attachment2a.txt"))

	equal, diags := a.StringSemanticEquals(t.Context(), b)
	require.False(t, diags.HasError(), diags.Errors())
	assert.True(t, equal, "identical contents under different paths must not force replace")
}

func TestHashedFileSemanticEquals_DifferentContent(t *testing.T) {
	a := NewHashedFileValue(filepath.Join("..", "provider", "fixtures", "attachment1.txt"))
	b := NewHashedFileValue(filepath.Join("..", "provider", "fixtures", "attachment2b.txt"))

	equal, diags := a.StringSemanticEquals(t.Context(), b)
	require.False(t, diags.HasError(), diags.Errors())
	assert.False(t, equal, "different contents must force replace")
}

func TestFileMustBeReadable_MissingFile(t *testing.T) {
	req := validator.StringRequest{
		Path:        path.Root("file"),
		ConfigValue: types.StringValue("non-existent"),
	}
	resp := &validator.StringResponse{}
	fileMustBeReadable().ValidateString(t.Context(), req, resp)
	require.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "no such file or directory")
}

func TestHashedContentSemanticEquals_LegacySHA1(t *testing.T) {
	raw := "Hello, I'm a text attachment"
	digest, err := ContentSha1Sum(raw)
	require.NoError(t, err)

	legacy := NewHashedContentValue(digest)
	configured := NewHashedContentValue(raw)

	equal, diags := legacy.StringSemanticEquals(t.Context(), configured)
	require.False(t, diags.HasError(), diags.Errors())
	assert.True(t, equal)
}
