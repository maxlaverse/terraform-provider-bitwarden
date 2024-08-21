package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	test_command "github.com/maxlaverse/terraform-provider-bitwarden/internal/command/test"
	"github.com/stretchr/testify/assert"
)

func TestProviderReauthenticateWithPasswordIfAuthenticatedOnDifferentServer(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status":                          `{"serverURL": "http://127.0.0.99/", "userEmail": "test@laverse.net", "status": "unlocked"}`,
		"logout":                          ``,
		"config server http://127.0.0.1/": ``,
		"login test@laverse.net --raw --passwordenv BW_PASSWORD": `session-key1234`,
	})
	defer removeMocks(t)

	providerConfiguration := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Configure(context.Background(), terraform.NewResourceConfigRaw(providerConfiguration))

	if !assert.False(t, diag.HasError()) {
		t.Fatalf("unexpected error: %v", diag[0])
	}

	assert.Equal(t, []string{
		"status",
		"logout",
		"config server http://127.0.0.1/",
		"login test@laverse.net --raw --passwordenv BW_PASSWORD",
	}, commandsExecuted())
}

func TestProviderReauthenticateWithPasswordIfAuthenticatedWithDifferentUser(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "as-an-other-user@laverse.net", "status": "unlocked"}`,
		"logout": ``,
		"login test@laverse.net --raw --passwordenv BW_PASSWORD": `session-key1234`,
	})
	defer removeMocks(t)

	providerConfiguration := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Configure(context.Background(), terraform.NewResourceConfigRaw(providerConfiguration))

	if !assert.False(t, diag.HasError()) {
		t.Fatalf("unexpected error: %v", diag[0])
	}

	assert.Equal(t, []string{
		"status",
		"logout",
		"login test@laverse.net --raw --passwordenv BW_PASSWORD",
	}, commandsExecuted())
}

func TestProviderDoesntLogoutFirstIfUnauthenticated(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "as-an-other-user@laverse.net", "status": "unauthenticated"}`,
		"login test@laverse.net --raw --passwordenv BW_PASSWORD": `session-key1234`,
	})
	defer removeMocks(t)

	providerConfiguration := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Configure(context.Background(), terraform.NewResourceConfigRaw(providerConfiguration))

	if !assert.False(t, diag.HasError()) {
		t.Fatalf("unexpected error: %v", diag[0])
	}

	assert.Equal(t, []string{
		"status",
		"login test@laverse.net --raw --passwordenv BW_PASSWORD",
	}, commandsExecuted())
}

func TestProviderReauthenticateWithAPIIfAuthenticatedWithDifferentUser(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status":                                 `{"serverURL": "http://127.0.0.1/", "userEmail": "as-an-other-user@laverse.net", "status": "unlocked"}`,
		"logout":                                 ``,
		"login --apikey":                         ``,
		"unlock --raw --passwordenv BW_PASSWORD": `session-key1234`,
		"sync":                                   ``,
	})
	defer removeMocks(t)

	providerConfiguration := map[string]interface{}{
		"server":          "http://127.0.0.1/",
		"email":           "test@laverse.net",
		"client_id":       "client-id-1234",
		"client_secret":   "client-secret-5678",
		"master_password": "master-password-9",
	}

	diag := New(versionDev)().Configure(context.Background(), terraform.NewResourceConfigRaw(providerConfiguration))

	if !assert.False(t, diag.HasError()) {
		t.Fatalf("unexpected error: %v", diag[0])
	}

	assert.Equal(t, []string{
		"status",
		"logout",
		"login --apikey",
		"unlock --raw --passwordenv BW_PASSWORD",
	}, commandsExecuted())
}

func TestProviderWithSessionKeySync(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "test@laverse.net", "status": "unlocked"}`,
		"sync":   ``,
	})
	defer removeMocks(t)

	raw := map[string]interface{}{
		"server":      "http://127.0.0.1/",
		"email":       "test@laverse.net",
		"session_key": "abcd1234",
	}

	// We specifically set the provider's version to something else than 'versionDev'
	// in order to capture 'sync' calls.
	diag := New("not-dev")().Configure(context.Background(), terraform.NewResourceConfigRaw(raw))
	if !assert.False(t, diag.HasError()) {
		t.Fatal(diag[0])
	}

	assert.Equal(t, []string{
		"status",
		"sync",
	}, commandsExecuted())
}

func TestProviderRetryOnRateLimitExceeded(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status @error": `Rate limit exceeded. Try again later.`,
	})
	defer removeMocks(t)

	raw := map[string]interface{}{
		"server":      "http://127.0.0.1/",
		"email":       "test@laverse.net",
		"session_key": "abcd1234",
	}

	diag := New(versionDev)().Configure(context.Background(), terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, diag[0].Summary, "failing command 'status' for test purposes: Rate limit exceeded. Try again later.")
		assert.Equal(t, []string{
			"status",
			"status",
			"status",
		}, commandsExecuted())
	}
}

func TestProviderReturnUnhandledError(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status @error": `Something unknown and bad happened.`,
	})
	defer removeMocks(t)

	raw := map[string]interface{}{
		"server":      "http://127.0.0.1/",
		"email":       "test@laverse.net",
		"session_key": "abcd1234",
	}

	diag := New(versionDev)().Configure(context.Background(), terraform.NewResourceConfigRaw(raw))

	if assert.True(t, diag.HasError()) {
		assert.Equal(t, diag[0].Summary, "failing command 'status' for test purposes: Something unknown and bad happened.")
		assert.Equal(t, []string{
			"status",
		}, commandsExecuted())
	}
}
