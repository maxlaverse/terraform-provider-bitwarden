package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func aes256Decode(cipherText []byte, encKey []byte, iv []byte) ([]byte, error) {

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("error creating new cipher block: %w", err)
	}

	plainText := make([]byte, len(cipherText))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, cipherText)

	return pkcs5Unpadding(plainText, block.BlockSize())
}

func aes256Encode(plainText []byte, key []byte, iv []byte, blockSize int) ([]byte, error) {
	plainTextPadded := pkcs5Padding([]byte(plainText), blockSize, len(plainText))

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
	paddingLen := int(src[srcLen-1])
	if paddingLen >= srcLen || paddingLen > blockSize {
		return nil, fmt.Errorf("bad padding size")
	}
	return src[:srcLen-paddingLen], nil
}

func pkcs5Padding(cipherText []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(cipherText)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padtext...)
}
