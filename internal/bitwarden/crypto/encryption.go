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

var (
	SafeMode = true
)

func EncryptAsString(plainValue []byte, key symmetrickey.Key) (string, error) {
	res, err := Encrypt(plainValue, key)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func Encrypt(plainValue []byte, key symmetrickey.Key) (*encryptedstring.EncryptedString, error) {
	if len(plainValue) == 0 {
		return nil, fmt.Errorf("trying to encrypt nothing")
	}
	randomIV := make([]byte, 16)
	_, err := rand.Read(randomIV)
	if err != nil {
		return nil, fmt.Errorf("error generating random bytes: %w", err)
	}

	if len(key.EncryptionKey) == 0 {
		return nil, fmt.Errorf("no encryption key was provided: %v", key.EncryptionKey)
	}
	if len(randomIV) != 16 {
		return nil, fmt.Errorf("bad IV length - expected 16, got: %d", len(randomIV))
	}

	data, err := aes256Encode(plainValue, key.EncryptionKey, randomIV, 16)
	if err != nil {
		return nil, fmt.Errorf("error to aes256encoding data: %w", err)
	}

	hmac := hmacSum(append(randomIV, data...), key.MacKey, sha256.New)

	res := encryptedstring.New(randomIV, data, hmac, key)

	if SafeMode {
		safeDecryptedValue, err := Decrypt(&res, &key)
		if err != nil {
			return nil, fmt.Errorf("error reversing decryption (safe mode): %w", err)
		}
		if !bytes.Equal(safeDecryptedValue, plainValue) {
			return nil, fmt.Errorf("failed to reverse decryption (safe mode)")
		}
	}

	return &res, nil
}

func DecryptPrivateKey(encryptedPrivateKeyStr string, encryptionKey symmetrickey.Key) (*rsa.PrivateKey, error) {
	encString, err := encryptedstring.NewFromEncryptedValue(encryptedPrivateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private key: %w", err)
	}

	decryptedPrivateKey, err := Decrypt(encString, &encryptionKey)
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
		decEncKey, err = Decrypt(encKeyCipher, &key)
		if err != nil {
			return nil, fmt.Errorf("error decrypting encryption key: %w", err)
		}
	} else if encKeyCipher.Key.EncryptionType == symmetrickey.AesCbc256_HmacSha256_B64 {
		newKey := key.StretchKey()

		decEncKey, err = Decrypt(encKeyCipher, &newKey)
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

func Decrypt(encString *encryptedstring.EncryptedString, key *symmetrickey.Key) ([]byte, error) {
	if encString.Key.EncryptionType == symmetrickey.AesCbc128_HmacSha256_B64 && key.EncryptionType == symmetrickey.AesCbc256_B64 {
		return nil, fmt.Errorf("unsupported old scheme")
	}

	if encString.Key.EncryptionType != key.EncryptionType {
		return nil, fmt.Errorf("bad encryption type: %d!=%d", encString.Key.EncryptionType, key.EncryptionType)
	}

	if len(encString.Hmac) == 0 && len(key.MacKey) > 0 {
		return nil, fmt.Errorf("hmac value is missing")
	}

	if len(encString.Hmac) != len(key.MacKey) {
		return nil, fmt.Errorf("hmac lengths differ: %d!=%d", len(encString.Hmac), len(key.MacKey))
	}

	computedHmac := hmacSum(append(append([]byte{}, encString.IV...), encString.Data...), key.MacKey, sha256.New)
	if !bytes.Equal(computedHmac, encString.Hmac) {
		return nil, fmt.Errorf("hmac comparison failed: %v != %v", computedHmac, encString.Hmac)
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
