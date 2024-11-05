package embedded

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
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

func TestCreateAccessTokenLoginMock(t *testing.T) {
	t.Skip()
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

	client := webapi.NewClient(ServerURL, testDeviceIdentifer, testDeviceVersion)
	err = client.RegisterUser(ctx, signupRequest)
	if err != nil && !strings.Contains(err.Error(), "Registration not allowed or user already exists") {
		t.Fatal(err)
	}

	httpClient := http.Client{
		Transport: &diskTransport{
			Prefix: mockName,
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
