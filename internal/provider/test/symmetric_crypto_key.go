package test

import "fmt"

type SymmetricCryptoKey struct {
	key     []byte
	encKey  []byte
	macKey  []byte
	encType encryptionType
}

type encryptionType int

const (
	AesCbc256_B64            encryptionType = 0
	AesCbc128_HmacSha256_B64 encryptionType = 1
	AesCbc256_HmacSha256_B64 encryptionType = 2
	Rsa2048_OaepSha1_B64     encryptionType = 4
)

func NewSymmetricCryptoKey(rawKey []byte, encType encryptionType) (*SymmetricCryptoKey, error) {
	if encType == -1 {
		if len(rawKey) == 32 {
			encType = AesCbc256_B64
		} else if len(rawKey) == 64 {
			encType = AesCbc256_HmacSha256_B64
		} else {
			return nil, fmt.Errorf("unsupported raw key len: %d", len(rawKey))
		}
	}

	key := SymmetricCryptoKey{
		key:     rawKey,
		encType: encType,
	}
	if encType == AesCbc256_B64 && len(rawKey) == 32 {
		key.encKey = rawKey
	} else if encType == AesCbc128_HmacSha256_B64 && len(rawKey) == 32 {
		key.encKey = rawKey[0:16]
		key.macKey = rawKey[16:]
	} else if encType == AesCbc256_HmacSha256_B64 && len(rawKey) == 64 {
		key.encKey = rawKey[0:32]
		key.macKey = rawKey[32:]
	} else {
		return nil, fmt.Errorf("unsupported encryption type and len: %d/%d", encType, len(rawKey))
	}

	return &key, nil
}
