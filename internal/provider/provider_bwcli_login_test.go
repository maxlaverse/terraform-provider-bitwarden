//go:build offline

package provider

import (
	"testing"

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

	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		Email:          "test@laverse.net",
		MasterPassword: "master-password-9",
	}

	_, err := configureClients(t.Context(), versionTestDisabledRetries, cfg)
	if !assert.NoError(t, err) {
		t.Fatalf("unexpected error: %v", err)
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

	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		Email:          "test@laverse.net",
		MasterPassword: "master-password-9",
	}

	_, err := configureClients(t.Context(), versionTestDisabledRetries, cfg)
	if !assert.NoError(t, err) {
		t.Fatalf("unexpected error: %v", err)
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

	cfg := providerConfig{
		Server:         "http://127.0.0.1/",
		Email:          "test@laverse.net",
		MasterPassword: "master-password-9",
	}

	_, err := configureClients(t.Context(), versionTestDisabledRetries, cfg)
	if !assert.NoError(t, err) {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, []string{
		"status",
		"login test@laverse.net --raw --passwordenv BW_PASSWORD",
	}, commandsExecuted())
}

func TestProviderWithSessionKeySync(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "test@laverse.net", "status": "unlocked"}`,
		"sync":   ``,
	})
	defer removeMocks(t)

	cfg := providerConfig{
		Server:     "http://127.0.0.1/",
		Email:      "test@laverse.net",
		SessionKey: "abcd1234",
	}

	// We specifically set the provider's version to something else than 'versionTestDisabledRetries'
	// in order to capture 'sync' calls.
	_, err := configureClients(t.Context(), "not-dev", cfg)
	if !assert.NoError(t, err) {
		t.Fatal(err)
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

	cfg := providerConfig{
		Server:     "http://127.0.0.1/",
		Email:      "test@laverse.net",
		SessionKey: "abcd1234",
	}

	_, err := configureClients(t.Context(), versionTestDisabledRetries, cfg)

	if assert.Error(t, err) {
		assert.Equal(t, "failing command 'status' for test purposes: Rate limit exceeded. Try again later.", err.Error())
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

	cfg := providerConfig{
		Server:     "http://127.0.0.1/",
		Email:      "test@laverse.net",
		SessionKey: "abcd1234",
	}

	_, err := configureClients(t.Context(), versionTestDisabledRetries, cfg)

	if assert.Error(t, err) {
		assert.Equal(t, "failing command 'status' for test purposes: Something unknown and bad happened.", err.Error())
		assert.Equal(t, []string{
			"status",
		}, commandsExecuted())
	}
}
