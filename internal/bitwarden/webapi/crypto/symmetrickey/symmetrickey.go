package symmetrickey

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/hkdf"
)

type Key struct {
	Key            []byte
	EncryptionKey  []byte
	EncryptionType EncryptionType
	MacKey         []byte
}

type EncryptionType int

const (
	AesCbc256_B64            EncryptionType = 0
	AesCbc128_HmacSha256_B64 EncryptionType = 1
	AesCbc256_HmacSha256_B64 EncryptionType = 2
	Rsa2048_OaepSha256_B64   EncryptionType = 3
	Rsa2048_OaepSha1_B64     EncryptionType = 4
)

func NewFromRawBytesWithEncryptionType(rawKey []byte, encType EncryptionType) (*Key, error) {
	key := Key{
		Key:            rawKey,
		EncryptionType: encType,
	}
	if encType == AesCbc256_B64 && len(rawKey) == 32 {
		key.EncryptionKey = rawKey
	} else if encType == AesCbc128_HmacSha256_B64 && len(rawKey) == 32 {
		key.EncryptionKey = rawKey[0:16]
		key.MacKey = rawKey[16:]
	} else if encType == AesCbc256_HmacSha256_B64 && len(rawKey) == 64 {
		key.EncryptionKey = rawKey[0:32]
		key.MacKey = rawKey[32:]
	} else {
		return nil, fmt.Errorf("unsupported encryption type and len: %d/%d", encType, len(rawKey))
	}

	return &key, nil
}

func NewFromRawBytes(rawKey []byte) (*Key, error) {
	if len(rawKey) == 32 {
		return NewFromRawBytesWithEncryptionType(rawKey, AesCbc256_B64)
	} else if len(rawKey) == 64 {
		return NewFromRawBytesWithEncryptionType(rawKey, AesCbc256_HmacSha256_B64)
	}
	return nil, fmt.Errorf("unsupported raw key len: %d", len(rawKey))
}

func (key *Key) StretchKey() (*Key, error) {
	encKey := hkdf.Expand(sha256.New, key.Key, []byte("enc"))
	macKey := hkdf.Expand(sha256.New, key.Key, []byte("mac"))

	newEncKey := make([]byte, 32)
	encKey.Read(newEncKey)

	newMacKey := make([]byte, 32)
	macKey.Read(newMacKey)

	newKey := append(newEncKey, newMacKey...)
	return NewFromRawBytes(newKey)
}
