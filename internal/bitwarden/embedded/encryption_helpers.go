package embedded

import (
	"crypto/rsa"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/encryptedstring"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
)

func decryptOrganizationKey(key string, RSAPrivateKey *rsa.PrivateKey) (*symmetrickey.Key, error) {
	devV, err := keybuilder.RSADecrypt(key, RSAPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting organization key: %w", err)
	}

	return symmetrickey.NewFromRawBytes(devV)
}

func decryptStringAsBytes(cipherText string, key symmetrickey.Key) ([]byte, error) {
	encStr, err := encryptedstring.NewFromEncryptedValue(cipherText)
	if err != nil {
		return nil, fmt.Errorf("error creating encrypted string from '%s': %v", cipherText, err)
	}

	return crypto.Decrypt(encStr, &key)
}

func decryptStringAsKey(cipherText string, key symmetrickey.Key) (*symmetrickey.Key, error) {
	decStr, err := decryptStringAsBytes(cipherText, key)
	if err != nil {
		return nil, fmt.Errorf("error decrypting string as bytes: %v", err)
	}

	return symmetrickey.NewFromRawBytes(decStr)
}

func decryptStringIfNotEmpty(input string, key symmetrickey.Key) (string, error) {
	if len(input) == 0 {
		return "", nil
	}
	decrypted, err := decryptStringAsBytes(input, key)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func encryptAsStringIfNotEmpty(plainValue string, key symmetrickey.Key) (string, error) {
	if len(plainValue) == 0 {
		return "", nil
	}
	res, err := crypto.Encrypt([]byte(plainValue), key)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}
