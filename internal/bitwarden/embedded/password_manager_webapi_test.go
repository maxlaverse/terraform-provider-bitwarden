package embedded

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded/fixtures"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/assert"
)

func TestLoginAsPasswordLoadsAccountInformationForPbkdf2(t *testing.T) {
	vault, reset := newMockedPasswordManager(fixtures.MockedClient(t, fixtures.Pdkdf2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithPassword(ctx, fixtures.Pdkdf2Email, fixtures.TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, fixtures.Pdkdf2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 600000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, fixtures.Pdkdf2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, fixtures.Pdkdf2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsAPILoadsAccountInformationForPbkdf2(t *testing.T) {
	vault, reset := newMockedPasswordManager(fixtures.MockedClient(t, fixtures.Pdkdf2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithAPIKey(ctx, fixtures.TestPassword, "user.aaf15bd1-4f51-4ba0-ade8-9dc2ec0fd2c3", "ZTXHHyPY6bNlNq1diDA2nM1GROboP3")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, fixtures.Pdkdf2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 600000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, fixtures.Pdkdf2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, fixtures.Pdkdf2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsPasswordLoadsAccountInformationForArgon2(t *testing.T) {
	vault, reset := newMockedPasswordManager(fixtures.MockedClient(t, fixtures.Argon2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithPassword(ctx, fixtures.Argon2Email, fixtures.TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, fixtures.Argon2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, fixtures.Argon2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, fixtures.Argon2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsAPILoadsAccountInformationForArgon2(t *testing.T) {
	vault, reset := newMockedPasswordManager(fixtures.MockedClient(t, fixtures.Argon2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithAPIKey(ctx, fixtures.TestPassword, "user.3f0abf17-e779-4312-a3dd-9c6266e95a9e", "oQAvXGx5h3iw0wzzgRwySsGxn3PvvA")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, fixtures.Argon2Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, fixtures.Argon2ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, fixtures.Argon2ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestObjectCreation(t *testing.T) {
	vault, reset := newMockedPasswordManager(fixtures.MockedClient(t, fixtures.Pdkdf2Mocks))
	defer reset()

	ctx := context.Background()
	err := vault.LoginWithPassword(ctx, fixtures.Pdkdf2Email, fixtures.TestPassword)
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
		serverURL: fixtures.ServerURL,
		client:    client,
	}, httpmock.DeactivateAndReset
}
