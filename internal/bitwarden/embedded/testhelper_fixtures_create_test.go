//go:build offline

package embedded

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/helpers"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

// This is only used to generate test data
func TestCreateTestAccounts(t *testing.T) {
	t.Skip("Skipping test as it's only used to generate test data")

	createTestAccount(t, Pdkdf2Mocks, models.KdfConfiguration{
		KdfType:       models.KdfTypePBKDF2_SHA256,
		KdfIterations: 1000,
	}, true)
	createTestAccount(t, Argon2Mocks, models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}, false)

	createOrganization(t, Pdkdf2Mocks, Argon2Mocks)
	createOrganizationResources(t, Pdkdf2Mocks, OrganizationID)
}

func TestCreateAccessTokenLoginMock(t *testing.T) {
	t.Skip("Skipping test as it's only used to generate test data")

	accessTokenEncryptionKey, err := base64.StdEncoding.DecodeString("dGVzdC1lbmNyeXB0aW9uLWtleQ==")
	if err != nil {
		t.Fatal(err)
	}

	extractedKey := helpers.HMACSum(accessTokenEncryptionKey, []byte("bitwarden-accesstoken"), sha256.New)
	expandedKey := helpers.HKDFExpand(extractedKey, []byte("sm-access-token"), sha256.New, 64)

	userEncryptionKey, err := symmetrickey.NewFromRawBytesWithEncryptionType(expandedKey, symmetrickey.AesCbc256_HmacSha256_B64)
	if err != nil {
		t.Fatal(err)
	}

	mainEncryptionKeyBytes := base64.StdEncoding.EncodeToString([]byte("test-keytest-keytest-keytest-key"))

	plainValue := fmt.Sprintf("{\"encryptionKey\": \"%s\"}", mainEncryptionKeyBytes)
	res, err := crypto.Encrypt([]byte(plainValue), *userEncryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, &MachineAccountClaims{
		Organization: "b1a4b97f-c75e-4901-b831-00912f3549a7",
	})

	accessToken, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	final := webapi.MachineTokenResponse{
		AccessToken:      accessToken,
		EncryptedPayload: res.String(),
		TokenType:        "Bearer",
		Scope:            "api.secrets",
		ExpireIn:         3600,
	}

	out, err := json.MarshalIndent(final, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("test-bws_POST_identity_connect_token.json", out, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func createOrganization(t *testing.T, account1 string, account2 string) {
	ctx := t.Context()

	accountEmail1 := fmt.Sprintf("%s@laverse.net", account1)
	httpClient := http.Client{
		Transport: &diskTransport{
			Prefix: path.Join(fixturesDir(t), account1),
		},
	}
	vault := NewPasswordManagerClient(ServerURL, testDeviceIdentifer, "dev", WithPasswordManagerHttpOptions(webapi.WithCustomClient(httpClient)))

	err := vault.LoginWithPassword(ctx, accountEmail1, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	OrganizationID, err = vault.CreateOrganization(ctx, "test-organization", "test-organization-default", accountEmail1)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	accountEmail2 := fmt.Sprintf("%s@laverse.net", account2)
	err = vault.InviteUser(ctx, OrganizationID, accountEmail2, models.OrgMemberRoleTypeAdmin)
	if err != nil {
		t.Fatal(err)
	}

	_, err = vault.ConfirmInvite(ctx, OrganizationID, accountEmail2)
	if err != nil {
		t.Fatal(err)
	}
}

func createOrganizationResources(t *testing.T, account1 string, orgId string) {
	ctx := t.Context()

	accountEmail1 := fmt.Sprintf("%s@laverse.net", account1)
	httpClient := http.Client{
		Transport: &diskTransport{
			Prefix: path.Join(fixturesDir(t), account1),
		},
	}
	vault := NewPasswordManagerClient(ServerURL, testDeviceIdentifer, "dev", WithPasswordManagerHttpOptions(webapi.WithCustomClient(httpClient)))

	err := vault.LoginWithPassword(ctx, accountEmail1, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	orgCol, err := vault.CreateOrganizationCollection(ctx, models.OrgCollection{
		Object:         models.ObjectTypeOrgCollection,
		Name:           "org-collection",
		OrganizationID: orgId,
		Users:          []models.OrgCollectionMember{},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = vault.CreateItem(ctx, models.Item{
		Object:         models.ObjectTypeItem,
		Type:           models.ItemTypeLogin,
		Name:           "Item in org Vault",
		OrganizationID: orgCol.OrganizationID,
		CollectionIds:  []string{orgCol.ID},
		Login: models.Login{
			Username: "my-org-username",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func createTestAccount(t *testing.T, mockName string, kdfConfig models.KdfConfiguration, withResources bool) {
	ctx := t.Context()

	accountEmail := fmt.Sprintf("%s@laverse.net", mockName)

	preloginKey, err := keybuilder.BuildPreloginKey(TestPassword, accountEmail, kdfConfig)
	if err != nil {
		t.Fatal(err)
	}

	hashedPassword := crypto.HashPassword(TestPassword, *preloginKey, false)

	block, _ := pem.Decode([]byte(RsaPrivateKey))
	if block == nil {
		t.Fatal(err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	encryptionKeyBytes, err := base64.StdEncoding.DecodeString(EncryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	newEncryptionKey, encryptedEncryptionKey, err := keybuilder.EncryptEncryptionKey(*preloginKey, encryptionKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, encryptedPrivateKey, err := keybuilder.EncryptRSAKeyPair(*newEncryptionKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	signupRequest := webapi.SignupRequest{
		Email:              accountEmail,
		Name:               accountEmail,
		MasterPasswordHash: hashedPassword,
		Key:                encryptedEncryptionKey,
		Kdf:                kdfConfig.KdfType,
		KdfIterations:      kdfConfig.KdfIterations,
		KdfMemory:          kdfConfig.KdfMemory,
		KdfParallelism:     kdfConfig.KdfParallelism,
		Keys: webapi.KeyPair{
			PublicKey:           publicKey,
			EncryptedPrivateKey: encryptedPrivateKey,
		},
	}

	client := webapi.NewClient(ServerURL, testDeviceIdentifer, testDeviceVersion)
	err = client.RegisterUser(ctx, signupRequest)
	if err != nil && !strings.Contains(err.Error(), "Registration not allowed or user already exists") {
		t.Fatal(err)
	}

	httpClient := http.Client{
		Transport: &diskTransport{
			Prefix: path.Join(fixturesDir(t), mockName),
		},
	}
	vault := NewPasswordManagerClient(ServerURL, testDeviceIdentifer, "dev", WithPasswordManagerHttpOptions(webapi.WithCustomClient(httpClient)))

	err = vault.LoginWithPassword(ctx, accountEmail, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	apiKey, err := vault.GetAPIKey(ctx, accountEmail, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	err = vault.Logout(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = vault.LoginWithAPIKey(ctx, TestPassword, apiKey.ClientID, apiKey.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}

	if !withResources {
		return
	}

	_, err = vault.CreateItem(ctx, models.Item{
		Object: models.ObjectTypeItem,
		Type:   models.ItemTypeLogin,
		Name:   "Item in own Vault",
		Login: models.Login{
			Username: "my-username",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = vault.CreateFolder(ctx, models.Folder{
		Object: models.ObjectTypeFolder,
		Name:   "Folder in own Vault",
	})
	if err != nil {
		t.Fatal(err)
	}
}
