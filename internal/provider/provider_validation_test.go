package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestProviderSchemaValidity(t *testing.T) {
	if err := New(versionDev)().InternalValidate(); err != nil {
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

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	assert.False(t, diag.HasError())
}

func TestProviderAuthAPIMethodMissingClientIDThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Contains(t, diag[0].Detail, "all of `client_id,client_secret,master_password` must be specified")
	}
}

func TestProviderAuthAPIMethodMissingClientSecretThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_id":       "client-id-1234",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Regexp(t, regexp.MustCompile("all of `client_id,client_secret,master_password` must be specified|one of `master_password,session_key` must be specified"), diag[0].Detail)
	}
}

func TestProviderAuthAPIMethodMissingMasterPasswordThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":        "http://127.0.0.1/",
		"email":         "test@laverse.net",
		"client_id":     "client-id-1234",
		"client_secret": "client-secret-5678",
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Regexp(t, regexp.MustCompile("all of `client_id,client_secret,master_password` must be specified|one of `master_password,session_key` must be specified"), diag[0].Detail)
	}
}

func TestProviderAuthPasswordMethodMissingMasterPasswordThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server": "http://127.0.0.1/",
		"email":  "test@laverse.net",
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "Missing required argument", diag[0].Summary)
		assert.Regexp(t, "\"(master_password|session_key)\": one of `master_password,session_key` must be specified", diag[0].Detail)
	}
}

func TestProviderAuthSessionMethodValid(t *testing.T) {
	raw := map[string]interface{}{
		"server":      "http://127.0.0.1/",
		"email":       "test@laverse.net",
		"session_key": "1234",
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	assert.False(t, diag.HasError())
}

func TestProviderAuthAllMethodsMissingEmailThrowsError(t *testing.T) {
	raw := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

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

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	assert.False(t, diag.HasError())
}

func TestSessionKeyWithEmbeddedClientFails(t *testing.T) {
	raw := map[string]interface{}{
		"email":       "test@laverse.net",
		"session_key": "test1234",
		"experimental": []interface{}{
			map[string]interface{}{
				"embedded_client": true,
			},
		},
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "\"session_key\": conflicts with experimental.0.embedded_client", diag[0].Detail)
	}
}

func TestExperimentBoolRaiseIfFalse(t *testing.T) {
	raw := map[string]interface{}{
		"email":           "test@laverse.net",
		"master_password": "master-password-9",
		"experimental": []interface{}{
			map[string]interface{}{
				"embedded_client": false,
			},
		},
	}

	diag := New(versionDev)().Validate(terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, "\"experimental.0.embedded_client\": can't be false if set", diag[0].Summary)
	}
}
