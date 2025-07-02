//go:build offline

package embedded

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/assert"
)

func TestLoginAsPasswordLoadsAccountInformationForPbkdf2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithPassword(ctx, AccountPbkdf2.Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, AccountPbkdf2.Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 1000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, AccountPbkdf2.ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, AccountPbkdf2.ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsAPILoadsAccountInformationForPbkdf2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithAPIKey(ctx, TestPassword, "user.aaf15bd1-4f51-4ba0-ade8-9dc2ec0fd2c3", "ZTXHHyPY6bNlNq1diDA2nM1GROboP3")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, AccountPbkdf2.Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 1000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, AccountPbkdf2.ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, AccountPbkdf2.ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsPasswordLoadsAccountInformationForArgon2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Argon2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithPassword(ctx, AccountArgon2.Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, AccountArgon2.Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, AccountArgon2.ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, AccountArgon2.ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsAPILoadsAccountInformationForArgon2(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Argon2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithAPIKey(ctx, TestPassword, "user.3f0abf17-e779-4312-a3dd-9c6266e95a9e", "oQAvXGx5h3iw0wzzgRwySsGxn3PvvA")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, AccountArgon2.Email, vault.loginAccount.Email)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, AccountArgon2.ProtectedRSAPrivateKey, vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, AccountArgon2.ProtectedSymmetricKey, vault.loginAccount.ProtectedSymmetricKey)
}

func TestFolderCreation(t *testing.T) {
	// Goal is to test end-to-end creation but most importantly to have another
	// layer of detection for sensitive information not being encrypted.
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithPassword(ctx, AccountPbkdf2.Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	newObj := testFullyFilledFolder()

	obj, err := vault.CreateFolder(ctx, newObj)
	assert.NoError(t, err)
	if !assert.NotNil(t, obj) {
		return
	}

	assert.Equal(t, "Folder in own Vault", obj.Name)
}

func TestItemCreation(t *testing.T) {
	// Goal is to test end-to-end creation but most importantly to have another
	// layer of detection for sensitive information not being encrypted.
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithPassword(ctx, AccountPbkdf2.Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	newObj := testFullyFilledItem()
	newObj.OrganizationID = ""
	newObj.CollectionIds = nil

	obj, err := vault.CreateItem(ctx, newObj)
	assert.NoError(t, err)
	if !assert.NotNil(t, obj) {
		return
	}

	assert.Equal(t, "Item in own Vault", obj.Name)
	assert.Equal(t, "my-username", obj.Login.Username)
}

func TestItemCreationInOrganization(t *testing.T) {
	// Goal is to test end-to-end creation but most importantly to have another
	// layer of detection for sensitive information not being encrypted.
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	ctx := t.Context()
	err := vault.LoginWithPassword(ctx, AccountPbkdf2.Email, TestPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	vault.Sync(ctx)
	newObj := testFullyFilledItem()
	newObj.OrganizationID = OrganizationID
	newObj.CollectionIds = []string{"simply-not-empty"}

	obj, err := vault.CreateItem(ctx, newObj)
	assert.NoError(t, err)
	if !assert.NotNil(t, obj) {
		return
	}

	assert.Equal(t, "Item in org Vault", obj.Name)
	assert.Equal(t, "my-org-username", obj.Login.Username)
}

func TestVerifyObjectAfterWrite(t *testing.T) {
	vault, reset := newMockedPasswordManager(MockedClient(t, Pdkdf2Mocks))
	defer reset()

	vault.failOnSyncAfterWriteVerification = true
	err := vault.verifyObjectAfterWrite(t.Context(), testFullyFilledItem(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server returned different models.Item after write!\nAfter writing a models.Item and re-fetching it, the server returned a slightly different version: different keys at []\n\nTo learn more about this issue and how to handle it, please:\n1. Consider reporting affected fields at: https://github.com/maxlaverse/terraform-provider-bitwarden/issues/new\n2. Check the documentation of the 'experimental.disable_sync_after_write_verification' attribute")
}

func newMockedPasswordManager(client webapi.Client) (webAPIVault, func()) {
	httpmock.Activate()

	return webAPIVault{
		serverURL: ServerURL,
		client:    client,
	}, httpmock.DeactivateAndReset
}
