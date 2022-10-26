package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestProviderSchemaValidity(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderAuthAPIMethodValid(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_id":       "client-id-1234",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	assert.False(t, diag.HasError())
}

func TestProviderAuthAPIMethodMissingClientIDThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Equal(t, "\"client_secret\": all of `client_id,client_secret` must be specified", diag[0].Detail)
	}
}

func TestProviderAuthAPIMethodMissingClientSecretThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_id":       "client-id-1234",
		"master_password": "master-password-9",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Equal(t, "\"client_id\": all of `client_id,client_secret` must be specified", diag[0].Detail)
	}
}

func TestProviderAuthAPIMethodMissingMasterPasswordThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":        "http://127.0.0.1/",
		"email":         "test@laverse.net",
		"client_id":     "client-id-1234",
		"client_secret": "client-secret-5678",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Equal(t, "The argument \"master_password\" is required, but no definition was found.", diag[0].Detail)
	}
}

func TestProviderAuthPasswordMethodMissingMasterPasswordThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server": "http://127.0.0.1/",
		"email":  "test@laverse.net",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Equal(t, "The argument \"master_password\" is required, but no definition was found.", diag[0].Detail)
	}
}

func TestProviderAuthSessionMethodValid(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"session_key":     "1234",
		"master_password": "master-password-9",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	assert.False(t, diag.HasError())
}

func TestProviderAuthAllMethodsMissingEmailThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"master_password": "master-password-9",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Equal(t, "The argument \"email\" is required, but no definition was found.", diag[0].Detail)
	}
}

func TestProviderAuthAllMethodsMissingServerNoError(t *testing.T) {
	raw := map[string]interface{}{
		"email":           "test@laverse.net",
		"master_password": "master-password-9",
	}

	diag := New("dev")().Validate(terraform.NewResourceConfigRaw(raw))

	assert.False(t, diag.HasError())
}
