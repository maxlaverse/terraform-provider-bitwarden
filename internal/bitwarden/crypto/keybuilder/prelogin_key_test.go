//go:build offline

package keybuilder

import (
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/stretchr/testify/assert"
)

func TestArgon2(t *testing.T) {
	key, err := buildKey("test1234", "somesalt", models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Key: CAgX8/OUnQXSAzipUdqaQ9CFCflNf2lowXvfbpzNmXU=\nEncryptionKey: CAgX8/OUnQXSAzipUdqaQ9CFCflNf2lowXvfbpzNmXU=\nMacKey: \n", key.Summary())
}

func TestArgon2WithTooLittleParallelism(t *testing.T) {
	key, err := buildKey("test1234", "somesalt", models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 0,
	})

	assert.ErrorContains(t, err, "Argon2 parallelism must be between")
	assert.Nil(t, key)
}

func TestBuildKeyRejectsOutOfRangeKdf(t *testing.T) {
	tests := []struct {
		name   string
		config models.KdfConfiguration
		want   string
	}{
		{
			name:   "pbkdf2 zero",
			config: models.KdfConfiguration{KdfType: models.KdfTypePBKDF2_SHA256, KdfIterations: 0},
			want:   "PBKDF2 iterations",
		},
		{
			name:   "pbkdf2 negative",
			config: models.KdfConfiguration{KdfType: models.KdfTypePBKDF2_SHA256, KdfIterations: -1},
			want:   "PBKDF2 iterations",
		},
		{
			name:   "pbkdf2 below prelogin min",
			config: models.KdfConfiguration{KdfType: models.KdfTypePBKDF2_SHA256, KdfIterations: 1000},
			want:   "PBKDF2 iterations",
		},
		{
			name:   "pbkdf2 over max",
			config: models.KdfConfiguration{KdfType: models.KdfTypePBKDF2_SHA256, KdfIterations: 2_000_001},
			want:   "PBKDF2 iterations",
		},
		{
			name:   "argon2 iterations low",
			config: models.KdfConfiguration{KdfType: models.KdfTypeArgon2, KdfIterations: 1, KdfMemory: 64, KdfParallelism: 4},
			want:   "Argon2 iterations",
		},
		{
			name:   "argon2 iterations high",
			config: models.KdfConfiguration{KdfType: models.KdfTypeArgon2, KdfIterations: 11, KdfMemory: 64, KdfParallelism: 4},
			want:   "Argon2 iterations",
		},
		{
			name:   "argon2 memory low",
			config: models.KdfConfiguration{KdfType: models.KdfTypeArgon2, KdfIterations: 3, KdfMemory: 15, KdfParallelism: 4},
			want:   "Argon2 memory",
		},
		{
			name:   "argon2 memory high",
			config: models.KdfConfiguration{KdfType: models.KdfTypeArgon2, KdfIterations: 3, KdfMemory: 1025, KdfParallelism: 4},
			want:   "Argon2 memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := buildKey("test1234", "somesalt", tt.config)
			assert.ErrorContains(t, err, tt.want)
			assert.Nil(t, key)
		})
	}
}

func TestBuildKeyAcceptsPreloginMin(t *testing.T) {
	_, err := buildKey("test1234", "somesalt", models.KdfConfiguration{
		KdfType:       models.KdfTypePBKDF2_SHA256,
		KdfIterations: 5000,
	})
	assert.NoError(t, err)
}
