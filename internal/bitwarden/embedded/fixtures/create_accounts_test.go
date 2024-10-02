package fixtures

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"strings"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

// This is only used to generate test data
func TestCreateTestAccounts(t *testing.T) {
	t.Skip()
	createTestAccount(t, Pdkdf2Email, models.KdfConfiguration{
		KdfType:       models.KdfTypePBKDF2_SHA256,
		KdfIterations: 600000,
	})
	createTestAccount(t, Argon2Email, models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	})
}

func createTestAccount(t *testing.T, accountEmail string, kdfConfig models.KdfConfiguration) {
	ctx := context.Background()

	mockName := strings.Split(accountEmail, "@")[0]

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

	client := webapi.NewClient(ServerURL)
	err = client.RegisterUser(ctx, signupRequest)
	if err != nil && !strings.Contains(err.Error(), "Registration not allowed or user already exists") {
		t.Fatal(err)
	}

	httpClient := http.Client{
		Transport: &diskTransport{
			Prefix: mockName,
		},
	}
	vault := embedded.NewWebAPIVault(ServerURL, embedded.WithHttpOptions(webapi.WithCustomClient(httpClient)))

	err = vault.LoginWithPassword(ctx, accountEmail, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	apiKey, err := vault.GetAPIKey(ctx, accountEmail, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	err = vault.LoginWithAPIKey(ctx, TestPassword, apiKey.ClientID, apiKey.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}

	_, err = vault.CreateObject(ctx, models.Object{
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
}
