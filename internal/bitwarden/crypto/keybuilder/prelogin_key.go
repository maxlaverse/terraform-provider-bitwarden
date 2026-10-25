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

const (
	// Bitwarden setting-time PBKDF2 is 600_000–2_000_000; prelogin min is 5000.
	// The max is the DoS control against a malicious or compromised server.
	pbkdf2IterationsMin = 5000
	pbkdf2IterationsMax = 2_000_000
	argon2IterationsMin = 2
	argon2IterationsMax = 10
	argon2MemoryMinMiB  = 16
	argon2MemoryMaxMiB  = 1024
	argon2ParallelMin   = 1
	argon2ParallelMax   = 16
)

func buildKey(masterPassword, salt string, kdfConfig models.KdfConfiguration) (*symmetrickey.Key, error) {
	if err := validateKdfConfig(kdfConfig); err != nil {
		return nil, err
	}

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

func validateKdfConfig(kdfConfig models.KdfConfiguration) error {
	switch kdfConfig.KdfType {
	case models.KdfTypePBKDF2_SHA256:
		if kdfConfig.KdfIterations < pbkdf2IterationsMin || kdfConfig.KdfIterations > pbkdf2IterationsMax {
			return fmt.Errorf("PBKDF2 iterations must be between %d and %d, got %d", pbkdf2IterationsMin, pbkdf2IterationsMax, kdfConfig.KdfIterations)
		}
	case models.KdfTypeArgon2:
		if kdfConfig.KdfIterations < argon2IterationsMin || kdfConfig.KdfIterations > argon2IterationsMax {
			return fmt.Errorf("Argon2 iterations must be between %d and %d, got %d", argon2IterationsMin, argon2IterationsMax, kdfConfig.KdfIterations)
		}
		if kdfConfig.KdfMemory < argon2MemoryMinMiB || kdfConfig.KdfMemory > argon2MemoryMaxMiB {
			return fmt.Errorf("Argon2 memory must be between %d and %d MiB, got %d", argon2MemoryMinMiB, argon2MemoryMaxMiB, kdfConfig.KdfMemory)
		}
		if kdfConfig.KdfParallelism < argon2ParallelMin || kdfConfig.KdfParallelism > argon2ParallelMax {
			return fmt.Errorf("Argon2 parallelism must be between %d and %d, got %d", argon2ParallelMin, argon2ParallelMax, kdfConfig.KdfParallelism)
		}
	default:
		return fmt.Errorf("unsupported KDF: '%d'", kdfConfig.KdfType)
	}
	return nil
}
