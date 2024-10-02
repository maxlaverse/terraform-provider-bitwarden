package keybuilder

import (
	"crypto/sha256"
	"fmt"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
)

func BuildPreloginKey(masterPassword, email string, kdfConfig models.KdfConfiguration) (*symmetrickey.Key, error) {
	return buildKey(masterPassword, email, kdfConfig)
}

func buildKey(masterPassword, salt string, kdfConfig models.KdfConfiguration) (*symmetrickey.Key, error) {
	var rawKey []byte
	switch kdfConfig.KdfType {
	case models.KdfTypePBKDF2_SHA256:
		rawKey = pbkdf2.Key([]byte(masterPassword), []byte(salt), kdfConfig.KdfIterations, 32, sha256.New)
	case models.KdfTypeArgon2:
		hashedSalt := sha256.New()
		hashedSalt.Write([]byte(salt))

		var err error
		err = func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("argon2.IDKey() panicked: %v", r)
				}
			}()

			rawKey = argon2.IDKey([]byte(masterPassword), hashedSalt.Sum(nil), uint32(kdfConfig.KdfIterations), uint32(kdfConfig.KdfMemory*1024), uint8(kdfConfig.KdfParallelism), 32)
			return
		}()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported KDF: '%d'", kdfConfig.KdfType)
	}

	return symmetrickey.NewFromRawBytes(rawKey)
}
