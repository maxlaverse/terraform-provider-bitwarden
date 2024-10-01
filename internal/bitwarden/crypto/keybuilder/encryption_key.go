package keybuilder

import (
	"crypto/rand"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
)

func GenerateEncryptionKey(key symmetrickey.Key) (*symmetrickey.Key, string, error) {
	encryptionKey := make([]byte, 64)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		return nil, "", fmt.Errorf("error generating random bytes: %w", err)
	}

	return EncryptEncryptionKey(key, encryptionKey)
}

func EncryptEncryptionKey(key symmetrickey.Key, encryptionKey []byte) (newEncryptionKey *symmetrickey.Key, encryptedEncryptionKey string, err error) {
	if len(key.Key) == 32 {
		stretchedKey := key.StretchKey()
		encryptedEncryptionKey, err = crypto.EncryptAsString(encryptionKey, stretchedKey)
		if err != nil {
			return nil, "", fmt.Errorf("error encrypting encryption key (symmetric key len: %d): %w", len(key.Key), err)
		}
	} else if len(key.Key) == 64 {
		encryptedEncryptionKey, err = crypto.EncryptAsString(encryptionKey, key)
		if err != nil {
			return nil, "", fmt.Errorf("error encrypting encryption key (symmetric key  len: %d): %w", len(key.Key), err)
		}
	} else {
		return nil, "", fmt.Errorf("invalid symmetric key length %d", len(key.Key))
	}

	newEncryptionKey, err = symmetrickey.NewFromRawBytes(encryptionKey)
	return newEncryptionKey, encryptedEncryptionKey, err
}
