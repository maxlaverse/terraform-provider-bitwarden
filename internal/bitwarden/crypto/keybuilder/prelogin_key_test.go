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

	assert.Errorf(t, err, "panicked: argon2: parallelism degree too low")
	assert.Nil(t, key)
}
