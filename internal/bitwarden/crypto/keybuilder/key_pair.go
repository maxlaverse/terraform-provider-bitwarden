package keybuilder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
)

func GenerateKeyPair(key symmetrickey.Key) (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("error generating rsa key: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("error marshalling PKIX public key: %w", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("error marshalling PKIX private key: %w", err)
	}

	encryptedPrivateKey, err := crypto.EncryptAsString(privateKeyBytes, key)
	if err != nil {
		return "", "", fmt.Errorf("error encrypting private key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(publicKeyBytes), encryptedPrivateKey, nil
}
