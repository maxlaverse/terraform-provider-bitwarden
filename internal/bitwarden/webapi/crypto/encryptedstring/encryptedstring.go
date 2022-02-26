package encryptedstring

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi/crypto/symmetrickey"
)

type EncryptedString struct {
	IV   []byte
	Data []byte
	Hmac []byte
	Key  symmetrickey.Key
}

func New(iv, data, hmac []byte, key symmetrickey.Key) EncryptedString {
	return EncryptedString{
		IV:   iv,
		Key:  key,
		Data: data,
		Hmac: hmac,
	}
}

func NewFromEncryptedValue(encryptedValue string) (*EncryptedString, error) {
	var encPieces []string
	encString := EncryptedString{}

	headerPieces := strings.Split(encryptedValue, ".")
	if len(headerPieces) == 2 {
		s, err := strconv.ParseInt(headerPieces[0], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("unable to parse encryption type: %w", err)
		}
		encString.Key.EncryptionType = symmetrickey.EncryptionType(s)
		encPieces = strings.Split(headerPieces[1], "|")
	} else {
		encPieces = strings.Split(encryptedValue, "|")
		if len(encPieces) == 3 {
			encString.Key.EncryptionType = symmetrickey.AesCbc128_HmacSha256_B64
		} else {
			encString.Key.EncryptionType = symmetrickey.AesCbc256_B64
		}
	}

	switch encString.Key.EncryptionType {
	case symmetrickey.AesCbc128_HmacSha256_B64, symmetrickey.AesCbc256_HmacSha256_B64:
		if len(encPieces) != 3 {
			return nil, fmt.Errorf("bad length (expected: 3)")
		}

		encString.IV = []byte(encPieces[0])
		encString.Data = []byte(encPieces[1])
		encString.Hmac = []byte(encPieces[2])
	case symmetrickey.AesCbc256_B64:
		if len(encPieces) != 2 {
			return nil, fmt.Errorf("bad length (expected: 2)")
		}

		encString.IV = []byte(encPieces[0])
		encString.Data = []byte(encPieces[1])
	case symmetrickey.Rsa2048_OaepSha256_B64, symmetrickey.Rsa2048_OaepSha1_B64:
		if len(encPieces) != 1 {
			return nil, fmt.Errorf("bad length (expected: 1)")
		}

		encString.Data = []byte(encPieces[0])
	default:
		return nil, fmt.Errorf("unsupported encryption type)")
	}

	base64DecodedIV, err := base64.StdEncoding.DecodeString(string(encString.IV))
	if err != nil {
		return nil, fmt.Errorf("unable to base64 decode IV: %w", err)
	}

	base64DecodedData, err := base64.StdEncoding.DecodeString(string(encString.Data))
	if err != nil {
		return nil, fmt.Errorf("unable to base64 decode data: %w", err)
	}

	base64DecodedMac, err := base64.StdEncoding.DecodeString(string(encString.Hmac))
	if err != nil {
		return nil, fmt.Errorf("unable to base64 decode hmac: %w", err)
	}

	encString.IV = base64DecodedIV
	encString.Data = base64DecodedData
	encString.Hmac = base64DecodedMac
	return &encString, nil
}

func (encString *EncryptedString) String() string {
	base64EncodedIV := base64.StdEncoding.EncodeToString(encString.IV)
	base64EncodedData := base64.StdEncoding.EncodeToString(encString.Data)
	base64EncodedHmac := base64.StdEncoding.EncodeToString(encString.Hmac)

	encType := fmt.Sprintf("%d", encString.Key.EncryptionType)

	var encryptedString string
	if len(encString.IV) > 0 {
		encryptedString = encType + "." + base64EncodedIV + "|" + base64EncodedData
	} else {
		encryptedString = encType + "." + base64EncodedData
	}

	if len(encString.Hmac) > 0 {
		encryptedString = encryptedString + "|" + base64EncodedHmac
	}

	return encryptedString
}
