package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func aes256Decode(cipherText []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher block: %w", err)
	}

	plainText := make([]byte, len(cipherText))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, cipherText)

	return pkcs5Unpadding(plainText, block.BlockSize())
}

func aes256Encode(plainText []byte, key []byte, iv []byte, blockSize int) ([]byte, error) {
	plainTextPadded := pkcs5Padding([]byte(plainText), blockSize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher block: %w", err)
	}

	cipherText := make([]byte, len(plainTextPadded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plainTextPadded)

	return cipherText, nil
}

func pkcs5Unpadding(src []byte, blockSize int) ([]byte, error) {
	srcLen := len(src)
	if srcLen == 0 || srcLen%blockSize != 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	paddingLen := int(src[srcLen-1])
	if paddingLen < 1 || paddingLen > blockSize || paddingLen > srcLen {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := 0; i < paddingLen; i++ {
		if src[srcLen-1-i] != byte(paddingLen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return src[:srcLen-paddingLen], nil
}

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := (blockSize - len(cipherText)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padtext...)
}
