package test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	b64 "encoding/base64"
	"fmt"
	"hash"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2_SHA256 = 0
)

type crypto struct {
	randomEncryptionKey []byte
	randomIV            string
}

func (c *crypto) MakeEncryptionKey(theKey SymmetricCryptoKey) (*SymmetricCryptoKey, []byte, error) {
	return c.buildEncryptionKey(theKey, c.randomEncryptionKey)
}

func (c *crypto) aesEncrypt(plainValue []byte, key SymmetricCryptoKey) (*Obj, error) {
	bIV, err := b64.StdEncoding.DecodeString(c.randomIV)
	if err != nil {
		return nil, fmt.Errorf("unable to based64 decode IV: %w", err)
	}
	if len(key.encKey) == 0 {
		return nil, fmt.Errorf("no encryption key was provided: %v", key.encKey)
	}
	if len(bIV) != 16 {
		return nil, fmt.Errorf("bad IV length - expected 16, got: %d", len(bIV))
	}

	data := AES256Encode(plainValue, key.encKey, bIV, 16)

	hmac := hmacSum(append(bIV, data...), key.macKey, sha256.New)
	return &Obj{
		IV:   bIV,
		Key:  key,
		Data: data,
		Mac:  hmac,
	}, nil
}

func (c *crypto) encrypt(plainValue []byte, key SymmetricCryptoKey) ([]byte, error) {
	encObj, err := c.aesEncrypt(plainValue, key)
	if err != nil {
		return nil, fmt.Errorf("error while aesEncrypting: %w", err)
	}
	iv := b64.StdEncoding.EncodeToString(encObj.IV)
	data := b64.StdEncoding.EncodeToString(encObj.Data)
	hmac := b64.StdEncoding.EncodeToString(encObj.Mac)

	var encryptedString string
	encType := fmt.Sprintf("%d", encObj.Key.encType)
	if len(encObj.IV) > 0 {
		encryptedString = encType + "." + iv + "|" + data
	} else {
		encryptedString = encType + "." + data
	}

	if len(encObj.Mac) > 0 {
		encryptedString = encryptedString + "|" + hmac
	}

	return []byte(encryptedString), nil
}

func (c *crypto) buildEncryptionKey(key SymmetricCryptoKey, encKey []byte) (*SymmetricCryptoKey, []byte, error) {
	var encKeyEnc []byte

	if len(key.key) == 32 {
		newKey, err := stretchKey(key)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to stretch key: %w", err)
		}
		encKeyEnc, err = c.encrypt(encKey, *newKey)
		if err != nil {
			return nil, nil, fmt.Errorf("unable encrypt key: %w", err)
		}
	} else if len(key.key) == 64 {
		var err error
		encKeyEnc, err = c.encrypt(encKey, key)
		if err != nil {
			return nil, nil, fmt.Errorf("unable encrypt key: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("invalid key length %d", len(key.key))
	}

	newSym, err := NewSymmetricCryptoKey(encKey, -1)
	return newSym, encKeyEnc, err
}

func (c *crypto) MakeKeyPair(key SymmetricCryptoKey, privateKey *rsa.PrivateKey) (*KeyPair, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("error marshalling PKIX public key: %w", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error marshalling PKIX private key: %w", err)
	}

	encryptedPrivateKey, err := c.encrypt(privateKeyBytes, key)
	if err != nil {
		return nil, fmt.Errorf("error encrypting private key: %w", err)
	}

	return &KeyPair{
		PublicKey:           b64.StdEncoding.EncodeToString(publicKeyBytes),
		EncryptedPrivateKey: string(encryptedPrivateKey),
	}, nil

}

func (c *crypto) MakePreloginKey(masterPassword, email string, kdfIteration int) (*SymmetricCryptoKey, error) {
	return makeKey(masterPassword, email, PBKDF2_SHA256, kdfIteration)
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AES256Encode(plaintext []byte, key []byte, iv []byte, blockSize int) []byte {
	plaintextPadded := PKCS5Padding([]byte(plaintext), blockSize, len(plaintext))

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, len(plaintextPadded))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plaintextPadded)
	return ciphertext
}

func hmacSum(value, key []byte, algo func() hash.Hash) []byte {
	mac := hmac.New(algo, key)
	mac.Write(value)
	return mac.Sum(nil)
}

func makeKey(masterPassword, salt string, kdf, iterations int) (*SymmetricCryptoKey, error) {
	return NewSymmetricCryptoKey(pbkdf2.Key([]byte(masterPassword), []byte(salt), iterations, 32, sha256.New), -1)
}

func hashPassword(password string, key SymmetricCryptoKey, localAuthorization bool) string {
	iterations := 1
	if localAuthorization {
		iterations = 2
	}
	derivedKey := pbkdf2.Key(key.key, []byte(password), iterations, 32, sha256.New)
	return b64.StdEncoding.EncodeToString(derivedKey)
}

func stretchKey(key SymmetricCryptoKey) (*SymmetricCryptoKey, error) {
	encKey := hkdf.Expand(sha256.New, key.key, []byte("enc"))
	macKey := hkdf.Expand(sha256.New, key.key, []byte("mac"))

	newEncKey := make([]byte, 32)
	encKey.Read(newEncKey)

	newMacKey := make([]byte, 32)
	macKey.Read(newMacKey)

	newKey := append(newEncKey, newMacKey...)
	return NewSymmetricCryptoKey(newKey, -1)
}
