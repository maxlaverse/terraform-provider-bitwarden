package provider

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/stretchr/testify/assert"
)

func TestProviderSchemaValidity(t *testing.T) {
	if err := New(versionTestSkippedLogin)().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderAuthUsingAPIKey(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_id":       "client-id-1234",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
	}

	p := New(versionTestSkippedLogin)()

	config := terraform.NewResourceConfigRaw(raw)
	diag := p.Validate(config)
	assert.False(t, diag.HasError())

	diag = p.Configure(context.Background(), config)
	assert.False(t, diag.HasError())

	assert.Implements(t, (*bwcli.PasswordManagerClient)(nil), p.Meta())
}

func TestProviderAuthUsingAPIAndEmbedded(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_id":       "client-id-1234",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
		"experimental": []interface{}{
			map[string]interface{}{"embedded_client": "true"},
		},
	}

	p := New(versionTestSkippedLogin)()

	config := terraform.NewResourceConfigRaw(raw)
	diag := p.Validate(config)
	assert.False(t, diag.HasError())

	diag = p.Configure(context.Background(), config)
	assert.False(t, diag.HasError())

	assert.Implements(t, (*embedded.PasswordManagerClient)(nil), p.Meta())
}

func TestProviderAuthUsingSessionKey(t *testing.T) {
	raw := map[string]interface{}{
		"server":      "http://127.0.0.1/",
		"email":       "test@laverse.net",
		"session_key": "1234",
	}

	p := New(versionTestSkippedLogin)()

	config := terraform.NewResourceConfigRaw(raw)
	diag := p.Validate(config)
	if !assert.False(t, diag.HasError()) {
		t.Fatal(diag)
	}

	diag = p.Configure(context.Background(), config)
	assert.False(t, diag.HasError())

	assert.Implements(t, (*bwcli.PasswordManagerClient)(nil), p.Meta())
}

func TestProviderAuthUsingAccessToken(t *testing.T) {
	raw := map[string]interface{}{
		"access_token": "0.client_id.client_secret:dGVzdC1lbmNyeXB0aW9uLWtleQ==",
		"experimental": []interface{}{
			map[string]interface{}{"embedded_client": "true"},
		},
	}

	p := New(versionTestSkippedLogin)()

	config := terraform.NewResourceConfigRaw(raw)
	diag := p.Validate(config)
	if !assert.False(t, diag.HasError()) {
		t.Fatal(diag)
	}

	diag = p.Configure(context.Background(), config)
	if !assert.False(t, diag.HasError()) {
		t.Fatal(diag)
	}

	assert.Implements(t, (*embedded.SecretsManager)(nil), p.Meta())
}

func TestProviderAuthUsingAccessToken_ThrowsErrorWithoutExperimentalFlag(t *testing.T) {
	raw := map[string]interface{}{
		"access_token": "0.client_id.client_secret:dGVzdC1lbmNyeXB0aW9uLWtleQ==",
	}

	p := New(versionTestSkippedLogin)()

	config := terraform.NewResourceConfigRaw(raw)
	diag := p.Validate(config)
	if !assert.False(t, diag.HasError()) {
		t.Fatal(diag)
	}

	diag = p.Configure(context.Background(), config)
	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "access token is not supported without the experimental 'embedded_client' flag", diag[0].Summary)
	}
}

func TestProviderAuthUsingAPIKey_ThrowsErrorOnMissingClientID(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
	}

	diag := New(versionTestSkippedLogin)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
	}
}

func TestProviderAuthUsingAPIKey_ThrowsErrorOnMissingClientSecret(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"client_id":       "client-id-1234",
		"master_password": "master-password-9",
	}

	diag := New(versionTestSkippedLogin)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Regexp(t, regexp.MustCompile("all of `client_id,client_secret,master_password` must be specified|one of `master_password,session_key` must be specified"), diag[0].Detail)
	}
}

func TestProviderAuthUsingAPIKey_ThrowsErrorOnMissingMasterPassword(t *testing.T) {
	raw := map[string]interface{}{
		"server":        "http://127.0.0.1/",
		"email":         "test@laverse.net",
		"client_id":     "client-id-1234",
		"client_secret": "client-secret-5678",
	}

	diag := New(versionTestSkippedLogin)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Regexp(t, regexp.MustCompile("all of `client_id,client_secret,master_password` must be specified|one of `access_token,master_password,session_key` must be specified"), diag[0].Detail)
	}
}

func TestProviderAuthUsingUsernamePassword_ThrowsErrorOnMissingMasterPassword(t *testing.T) {
	raw := map[string]interface{}{
		"server": "http://127.0.0.1/",
		"email":  "test@laverse.net",
	}

	diag := New(versionTestSkippedLogin)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Regexp(t, "\"(master_password|session_key)\": one of `access_token,master_password,session_key` must be specified", diag[0].Detail)
	}
}

func TestProviderAuthForPasswordManager_ThrowsErrorOnMissingEmail(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"master_password": "master-password-9",
	}

	diag := New(versionTestSkippedLogin)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Equal(t, "\"email\": one of `access_token,client_id,email,session_key` must be specified", diag[0].Detail)
	}
}

func TestProviderAuth_ThrowsErrorOnMissingServer(t *testing.T) {
	raw := map[string]interface{}{
		"email":           "test@laverse.net",
		"master_password": "master-password-9",
	}

	p := New(versionTestSkippedLogin)()

	config := terraform.NewResourceConfigRaw(raw)
	diag := p.Validate(config)
	assert.False(t, diag.HasError())

	diag = p.Configure(context.Background(), config)
	assert.False(t, diag.HasError())

	assert.Implements(t, (*bwcli.PasswordManagerClient)(nil), p.Meta())
}
