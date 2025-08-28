package embedded

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAccessToken_ValidTokenStructure(t *testing.T) {
	// Test the specific access token provided by the user
	accessToken := "0.8802a9f2-4984-400f-998c-8074e848fea6.kIaAXm4uZwU6JafxEKLYLA==:Gh0MxDtCEP0tFL4+/R22Yg=="

	clientID, clientSecret, encryptionKey, err := parseAccessToken(accessToken)

	// Verify no error occurred
	assert.NoError(t, err)

	// Verify the parsed components
	assert.Equal(t, "8802a9f2-4984-400f-998c-8074e848fea6", clientID)
	assert.Equal(t, "kIaAXm4uZwU6JafxEKLYLA==", clientSecret)

	// Verify encryption key was created
	assert.NotNil(t, encryptionKey)

	// Verify the encryption key has the expected length (should be 64 bytes for AES-256-HMAC-SHA256)
	assert.Equal(t, 64, len(encryptionKey.Key))
}
