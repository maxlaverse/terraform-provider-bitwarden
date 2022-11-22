package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/executor"
	test_executor "github.com/maxlaverse/terraform-provider-bitwarden/internal/executor/test"
	"github.com/stretchr/testify/assert"
)

func TestProviderReauthenticateWithPasswordIfAuthenticatedOnDifferentServer(t *testing.T) {
	restoreDefaultExecutor := useFakeExecutor(t, map[string]string{
		"status":                          `{"serverURL": "http://127.0.0.99/", "userEmail": "test@laverse.net", "status": "unlocked"}`,
		"logout":                          ``,
		"config server http://127.0.0.1/": ``,
		"login test@laverse.net --raw --passwordenv BW_PASSWORD": `session-key1234`,
	})
	defer restoreDefaultExecutor(t)

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
	restoreDefaultExecutor := useFakeExecutor(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "as-an-other-user@laverse.net", "status": "unlocked"}`,
		"logout": ``,
		"login test@laverse.net --raw --passwordenv BW_PASSWORD": `session-key1234`,
	})
	defer restoreDefaultExecutor(t)

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
	restoreDefaultExecutor := useFakeExecutor(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "as-an-other-user@laverse.net", "status": "unauthenticated"}`,
		"login test@laverse.net --raw --passwordenv BW_PASSWORD": `session-key1234`,
	})
	defer restoreDefaultExecutor(t)

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
	restoreDefaultExecutor := useFakeExecutor(t, map[string]string{
		"status":                                 `{"serverURL": "http://127.0.0.1/", "userEmail": "as-an-other-user@laverse.net", "status": "unlocked"}`,
		"logout":                                 ``,
		"login --apikey":                         ``,
		"unlock --raw --passwordenv BW_PASSWORD": `session-key1234`,
		"sync":                                   ``,
	})
	defer restoreDefaultExecutor(t)

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
	restoreDefaultExecutor := useFakeExecutor(t, map[string]string{
		"status": `{"serverURL": "http://127.0.0.1/", "userEmail": "test@laverse.net", "status": "unlocked"}`,
		"sync":   ``,
	})
	defer restoreDefaultExecutor(t)

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

func useFakeExecutor(t *testing.T, dummyOutput map[string]string) func(t *testing.T) {
	old := executor.DefaultExecutor
	executor.DefaultExecutor = test_executor.New(dummyOutput)
	return func(t *testing.T) {
		executor.DefaultExecutor = old
	}
}

func commandsExecuted() []string {
	return executor.DefaultExecutor.(*test_executor.FakeExecutor).CommandsExecuted
}
