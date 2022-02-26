package keybuilder

import (
	"crypto/sha256"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi/crypto/symmetrickey"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2_SHA256 = 0
)

func BuildPreloginKey(masterPassword, email string, kdfIteration int) (*symmetrickey.Key, error) {
	return buildKey(masterPassword, email, PBKDF2_SHA256, kdfIteration)
}

func buildKey(masterPassword, salt string, kdf, iterations int) (*symmetrickey.Key, error) {
	return symmetrickey.NewFromRawBytes(pbkdf2.Key([]byte(masterPassword), []byte(salt), iterations, 32, sha256.New))
}
