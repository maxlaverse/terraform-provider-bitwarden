package keybuilder

import (
	"crypto/rand"
	"io"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
)

func CreateObjectKey() (*symmetrickey.Key, error) {
	objectKeyBytes := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, objectKeyBytes); err != nil {
		return nil, err
	}
	return symmetrickey.NewFromRawBytesWithEncryptionType(objectKeyBytes, symmetrickey.AesCbc256_HmacSha256_B64)
}
