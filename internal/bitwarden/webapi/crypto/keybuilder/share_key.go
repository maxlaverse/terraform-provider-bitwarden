package keybuilder

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi/crypto/symmetrickey"
)

func GenerateShareKey(publicKey *rsa.PublicKey) (string, *symmetrickey.Key, error) {
	shareKey := make([]byte, 64)
	rand.Read(shareKey)

	newKey, err := symmetrickey.NewFromRawBytes(shareKey)
	if err != nil {
		return "", nil, fmt.Errorf("error creating new symmetric crypto key")
	}

	encryptedShareKey, err := rsaEncrypt(shareKey, publicKey)
	if err != nil {
		return "", nil, fmt.Errorf("error encrypting share key")
	}

	return encryptedShareKey, newKey, nil
}

func rsaEncrypt(data []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha1.New(),
		rand.Reader,
		publicKey,
		[]byte(data),
		nil)
	if err != nil {
		return "", fmt.Errorf("error encrypting data using : %w", err)
	}

	return fmt.Sprintf("%d.%s", symmetrickey.Rsa2048_OaepSha1_B64, base64.StdEncoding.EncodeToString(encryptedBytes)), nil
}
