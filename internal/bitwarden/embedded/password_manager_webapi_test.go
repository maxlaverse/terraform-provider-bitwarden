package embedded

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/assert"
)

func TestLoginAsPasswordLoadsAccountInformationForPbkdf2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithPassword(ctx, Pdkdf2Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, Pdkdf2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 600000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, Pdkdf2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, Pdkdf2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsAPILoadsAccountInformationForPbkdf2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithAPIKey(ctx, TestPassword, "user.aaf15bd1-4f51-4ba0-ade8-9dc2ec0fd2c3", "ZTXHHyPY6bNlNq1diDA2nM1GROboP3")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, Pdkdf2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 600000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, Pdkdf2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, Pdkdf2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsPasswordLoadsAccountInformationForArgon2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Argon2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithPassword(ctx, Argon2Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, Argon2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, Argon2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, Argon2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsAPILoadsAccountInformationForArgon2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Argon2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithAPIKey(ctx, TestPassword, "user.3f0abf17-e779-4312-a3dd-9c6266e95a9e", "oQAvXGx5h3iw0wzzgRwySsGxn3PvvA")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, Argon2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, Argon2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, Argon2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestObjectCreation(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithPassword(ctx, Pdkdf2Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	obj, err := vault.CreateObject(ctx, models.Object{
		Object: models.ObjectTypeItem,
		Type:   models.ItemTypeLogin,
		Name:   "test",
	})
	assert.NoError(t, err)
	if !assert.NotNil(t, obj) {
		return
	}

	assert.Equal(t, "Item in own Vault", obj.Name)
	assert.Equal(t, "my-username", obj.Login.Username)
}

func newMockedPasswordManager(client webapi.Client) (webAPIVault, func()) {
	httpmock.Activate()

	return webAPIVault{
		serverURL: ServerURL,
		client:    client,
	}, httpmock.DeactivateAndReset
}
