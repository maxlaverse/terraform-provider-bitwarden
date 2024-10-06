package encryptedstring

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
)

const (
	BUFFER_ENC_TYPE_LENGTH = 1
	BUFFER_IV_LENGTH       = 16
	BUFFER_MAC_LENGTH      = 32
	BUFFER_MIN_DATA_LENGTH = 1
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

func NewFromEncryptedBuffer(encBytes []byte) (*EncryptedString, error) {
	encType := symmetrickey.EncryptionType(encBytes[0])

	encString := EncryptedString{}

	switch encType {
	case symmetrickey.AesCbc256_HmacSha256_B64:
		const minimumLength = BUFFER_ENC_TYPE_LENGTH + BUFFER_IV_LENGTH + BUFFER_MAC_LENGTH + BUFFER_MIN_DATA_LENGTH
		if len(encBytes) < minimumLength {
			return nil, fmt.Errorf("bad minimum length for encrypted buffer: %d", len(encBytes))
		}
		encString.Key.EncryptionType = encType
		encString.IV = encBytes[BUFFER_ENC_TYPE_LENGTH : BUFFER_ENC_TYPE_LENGTH+BUFFER_IV_LENGTH]
		encString.Hmac = encBytes[BUFFER_ENC_TYPE_LENGTH+BUFFER_IV_LENGTH : BUFFER_ENC_TYPE_LENGTH+BUFFER_IV_LENGTH+BUFFER_MAC_LENGTH]
		encString.Data = encBytes[BUFFER_ENC_TYPE_LENGTH+BUFFER_IV_LENGTH+BUFFER_MAC_LENGTH:]
	default:
		return nil, fmt.Errorf("unsupported encrypted buffer type: %d", encType)
	}

	return &encString, nil
}

func (encString *EncryptedString) ToEncryptedBuffer() ([]byte, error) {
	if len(encString.IV) != BUFFER_IV_LENGTH {
		return nil, fmt.Errorf("can't output encrypted buffer: bad IV length")
	}
	if len(encString.Hmac) != BUFFER_MAC_LENGTH {
		return nil, fmt.Errorf("can't output encrypted buffer: bad HMAC length")
	}
	if len(encString.Data) < BUFFER_MIN_DATA_LENGTH {
		return nil, fmt.Errorf("can't output encrypted buffer: bad data length")
	}
	encType := []byte{byte(encString.Key.EncryptionType)}
	encBuffer := append(encType, encString.IV...)
	encBuffer = append(encBuffer, encString.Hmac...)
	encBuffer = append(encBuffer, encString.Data...)

	return encBuffer, nil
}

func (encString *EncryptedString) Equals(otherString *EncryptedString) bool {
	if !bytes.Equal(encString.Data, otherString.Data) {
		return false
	}
	if !bytes.Equal(encString.IV, otherString.IV) {
		return false
	}
	if !bytes.Equal(encString.Hmac, otherString.Hmac) {
		return false
	}
	if encString.Key.EncryptionType != otherString.Key.EncryptionType {
		return false
	}
	return true
}

func NewFromEncryptedValue(encryptedValue string) (*EncryptedString, error) {
	if len(encryptedValue) == 0 {
		return nil, fmt.Errorf("supposedly encrypted string is empty")
	}
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
			return nil, fmt.Errorf("bad amount of pieces (expected: 3, got: %d)", len(encPieces))
		}

		encString.IV = []byte(encPieces[0])
		encString.Data = []byte(encPieces[1])
		encString.Hmac = []byte(encPieces[2])
	case symmetrickey.AesCbc256_B64:
		if len(encPieces) != 2 {
			return nil, fmt.Errorf("bad amount of pieces (expected: 2, got: %d)", len(encPieces))
		}

		encString.IV = []byte(encPieces[0])
		encString.Data = []byte(encPieces[1])
	case symmetrickey.Rsa2048_OaepSha256_B64, symmetrickey.Rsa2048_OaepSha1_B64:
		if len(encPieces) != 1 {
			return nil, fmt.Errorf("bad amount of pieces (expected: 1, got: %d)", len(encPieces))
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
