//go:build offline

package webapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccountUUID = "94901a39-c703-4a49-ad21-a442b2d74826"

func TestCreateItem_PersonalCipherJSONIncludesEncryptedFor(t *testing.T) {
	var got map[string]any
	c := newCipherJSONClient(t, &got)

	_, err := c.CreateItem(t.Context(), models.Item{
		Name: "personal",
		Type: models.ItemTypeLogin,
	})
	require.NoError(t, err)

	assert.Equal(t, testAccountUUID, got["encryptedFor"])
	_, hasKeyId := got["encryptedByKeyId"]
	assert.False(t, hasKeyId)
	_, hasNestedCipher := got["cipher"]
	assert.False(t, hasNestedCipher)
}

func TestCreateItem_OrganizationCipherJSONIncludesEncryptedFor(t *testing.T) {
	var got map[string]any
	c := newCipherJSONClient(t, &got)

	_, err := c.CreateItem(t.Context(), models.Item{
		Name:           "org",
		Type:           models.ItemTypeLogin,
		OrganizationID: "org-id",
		CollectionIds:  []string{"collection-id"},
	})
	require.NoError(t, err)

	cipher, ok := got["cipher"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, testAccountUUID, cipher["encryptedFor"])
	_, hasKeyId := cipher["encryptedByKeyId"]
	assert.False(t, hasKeyId)
	_, hasTopLevel := got["encryptedFor"]
	assert.False(t, hasTopLevel)
}

func TestEditItem_JSONIncludesEncryptedFor(t *testing.T) {
	var got map[string]any
	c := newCipherJSONClient(t, &got)

	_, err := c.EditItem(t.Context(), models.Item{
		ID:   "cipher-id",
		Name: "updated",
		Type: models.ItemTypeLogin,
	})
	require.NoError(t, err)

	assert.Equal(t, testAccountUUID, got["encryptedFor"])
	_, hasKeyId := got["encryptedByKeyId"]
	assert.False(t, hasKeyId)
}

func newCipherJSONClient(t *testing.T, got *map[string]any) Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, got))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cipher-id","object":"cipherDetails","type":1,"name":"n"}`))
	}))
	t.Cleanup(server.Close)

	c := NewClient(server.URL, "device-id", "dev", DisableRetries()).(*client)
	c.sessionAccessToken = testAccessToken(t, testAccountUUID)
	return c
}

func testAccessToken(t *testing.T, accountUUID string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: accountUUID})
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	return signed
}
