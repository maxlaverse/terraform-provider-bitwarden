package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"hash"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/encryptedstring"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
)

func Encrypt(plainValue []byte, key symmetrickey.Key) (string, error) {
	randomIV := make([]byte, 16)
	rand.Read(randomIV)

	if len(key.EncryptionKey) == 0 {
		return "", fmt.Errorf("no encryption key was provided: %v", key.EncryptionKey)
	}
	if len(randomIV) != 16 {
		return "", fmt.Errorf("bad IV length - expected 16, got: %d", len(randomIV))
	}

	data, err := aes256Encode(plainValue, key.EncryptionKey, randomIV, 16)
	if err != nil {
		return "", fmt.Errorf("error to aes256encoding data: %w", err)
	}

	hmac := hmacSum(append(randomIV, data...), key.MacKey, sha256.New)

	res := encryptedstring.New(randomIV, data, hmac, key)
	return res.String(), nil
}

func DecryptPrivateKey(encryptedPrivateKeyStr string, encryptionKey symmetrickey.Key) (*rsa.PrivateKey, error) {
	encString, err := encryptedstring.NewFromEncryptedValue(encryptedPrivateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private key: %w", err)
	}

	decryptedPrivateKey, err := decrypt(encString, &encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private key: %w", err)
	}

	p, err := x509.ParsePKCS8PrivateKey(decryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error parse private key: %w", err)
	}
	return p.(*rsa.PrivateKey), nil
}

func DecryptEncryptionKey(encryptedKeyStr string, key symmetrickey.Key) (*symmetrickey.Key, error) {
	var decEncKey []byte
	encKeyCipher, err := encryptedstring.NewFromEncryptedValue(encryptedKeyStr)
	if err != nil {
		return nil, fmt.Errorf("error decrypting encryption key: %w", err)
	}
	if encKeyCipher.Key.EncryptionType == symmetrickey.AesCbc256_B64 {
		decEncKey, err = decrypt(encKeyCipher, &key)
		if err != nil {
			return nil, fmt.Errorf("error decrypting encryption key: %w", err)
		}
	} else if encKeyCipher.Key.EncryptionType == symmetrickey.AesCbc256_HmacSha256_B64 {
		newKey, err := key.StretchKey()
		if err != nil {
			return nil, fmt.Errorf("error stretching key encryption key: %w", err)
		}

		decEncKey, err = decrypt(encKeyCipher, newKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting encryption key: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported encryption key type")
	}

	encryptionKey, err := symmetrickey.NewFromRawBytes(decEncKey)
	if err != nil {
		return nil, fmt.Errorf("unsupported decryption encryption key: %w", err)
	}
	return encryptionKey, nil
}

func decrypt(encString *encryptedstring.EncryptedString, key *symmetrickey.Key) ([]byte, error) {
	if encString.Key.EncryptionType == symmetrickey.AesCbc128_HmacSha256_B64 && key.EncryptionType == symmetrickey.AesCbc256_B64 {
		return nil, fmt.Errorf("unsupported old scheme")
	}

	if encString.Key.EncryptionType != key.EncryptionType {
		return nil, fmt.Errorf("bad encryption type: %d!=%d", encString.Key.EncryptionType, key.EncryptionType)
	}

	if len(encString.Hmac) == 0 && len(key.MacKey) > 0 {
		return nil, fmt.Errorf("hmac value is missing")
	}

	hmac := hmacSum(append(encString.IV, encString.Data...), key.MacKey, sha256.New)
	if !bytes.Equal(hmac, encString.Hmac) {
		return nil, fmt.Errorf("hmac comparison failed: %s!=%s", encString.Hmac, hmac)
	}
	decData, err := aes256Decode(encString.Data, key.EncryptionKey, encString.IV)
	if err != nil {
		return nil, fmt.Errorf("error aes256Decoding: %w", err)
	}
	return decData, nil
}

func hmacSum(value, key []byte, algo func() hash.Hash) []byte {
	mac := hmac.New(algo, key)
	mac.Write(value)
	return mac.Sum(nil)
}
