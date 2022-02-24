package crypto

import (
	"crypto/sha256"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi/crypto/symmetrickey"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/pbkdf2"
)

func TestHashPassword(t *testing.T) {
	preloginKey, err := symmetrickey.NewFromRawBytes(pbkdf2.Key([]byte("password"), []byte("salt"), 100000, 32, sha256.New))
	assert.NoError(t, err)

	hashedPassword := HashPassword("somepassword", *preloginKey, false)
	assert.Equal(t, "F0/RWOCogHqaZ17uC/aFv0XgDMUAAkAhQR99wbQQL2I=", hashedPassword)
}
