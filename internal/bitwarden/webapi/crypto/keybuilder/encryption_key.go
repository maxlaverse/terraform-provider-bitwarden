package keybuilder

import (
	"crypto/rand"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi/crypto/symmetrickey"
)

func GenerateEncryptionKey(key symmetrickey.Key) (*symmetrickey.Key, string, error) {
	encryptionKey := make([]byte, 64)
	rand.Read(encryptionKey)
	return buildEncryptionKey(key, encryptionKey)
}

func buildEncryptionKey(key symmetrickey.Key, encryptionKey []byte) (newEncryptionKey *symmetrickey.Key, encryptedEncryptionKey string, err error) {
	if len(key.Key) == 32 {
		stretchedKey, err := key.StretchKey()
		if err != nil {
			return nil, "", fmt.Errorf("error stretching key: %w", err)
		}
		encryptedEncryptionKey, err = crypto.Encrypt(encryptionKey, *stretchedKey)
		if err != nil {
			return nil, "", fmt.Errorf("error encrypting encryption key (symmetric key len: %d): %w", len(key.Key), err)
		}
	} else if len(key.Key) == 64 {
		encryptedEncryptionKey, err = crypto.Encrypt(encryptionKey, key)
		if err != nil {
			return nil, "", fmt.Errorf("error encrypting encryption key (symmetric key  len: %d): %w", len(key.Key), err)
		}
	} else {
		return nil, "", fmt.Errorf("invalid symmetric key length %d", len(key.Key))
	}

	newEncryptionKey, err = symmetrickey.NewFromRawBytes(encryptionKey)
	return newEncryptionKey, encryptedEncryptionKey, err
}
