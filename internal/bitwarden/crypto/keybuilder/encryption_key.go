package keybuilder

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/helpers"
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

func DeriveFromAccessTokenEncryptionKey(accessToken string) (*symmetrickey.Key, error) {
	accessTokenEncryptionKey, err := base64.StdEncoding.DecodeString(accessToken)
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding access token encryption key: %w", err)
	}

	extractedKey := helpers.HMACSum(accessTokenEncryptionKey, []byte("bitwarden-accesstoken"), sha256.New)
	expandedKey := helpers.HKDFExpand(extractedKey, []byte("sm-access-token"), sha256.New, 64)

	return symmetrickey.NewFromRawBytesWithEncryptionType(expandedKey, symmetrickey.AesCbc256_HmacSha256_B64)
}
