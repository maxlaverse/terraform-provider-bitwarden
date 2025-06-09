package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"bitwarden": func() (*schema.Provider, error) {
		version := versionTestDefault
		if !IsOfficialBackend() {
			version = versionTestDisabledRetries
		}
		return New(version)(), nil
	},
}

func ensureTestConfigurationReady(t *testing.T) {
	ensureTestAccountsExist(t)
	ensureTestResourcesExist(t)
}

func bwEmbeddedTestClient(t *testing.T) bitwarden.PasswordManager {
	client := embedded.NewPasswordManagerClient(testConfiguration.ServerURL, testDeviceIdentifer, testDeviceVersion)
	acc := testConfiguration.Accounts[testAccountFullAdmin]

	if err := client.LoginWithAPIKey(t.Context(), acc.Password, acc.ClientID, acc.ClientSecret); err != nil {
		t.Fatal(err)
	}
	t.Log("Logged in embedded test client")
	return client
}

func bwCLITestClient(t *testing.T) bwcli.PasswordManagerClient {
	vault, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	client := bwcli.NewPasswordManagerClient(bwcli.DisableRetryBackoff(), bwcli.WithAppDataDir(vault))
	status, err := client.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if len(status.ServerURL) == 0 {
		if err := client.SetServer(t.Context(), testConfiguration.ServerURL); err != nil {
			t.Fatal(err)
		}
	}

	acc := testConfiguration.Accounts[testAccountFullAdmin]
	switch status.Status {
	case bwcli.StatusUnauthenticated:
		if err := loginWithRetry(t, client, acc); err != nil {
			t.Fatal(err)
		}
	case bwcli.StatusLocked:
		if err := client.Unlock(t.Context(), acc.Password); err != nil {
			t.Fatal(err)
		}
	default:
		t.Logf("Official test client already logged-in: %s", status.Status)
	}
	return client
}

func loginWithRetry(t *testing.T, client bwcli.PasswordManagerClient, acc testAccount) error {
	for retries := 0; retries < 3; retries++ {
		if err := client.LoginWithAPIKey(t.Context(), acc.Password, acc.ClientID, acc.ClientSecret); err == nil {
			return nil
		}
		t.Log("Account creation not taken into account yet, retrying...")
		time.Sleep(time.Duration(retries+1) * time.Second)
	}
	return fmt.Errorf("failed to login after 3 retries")
}

func tfConfigPasswordManagerProvider(account testAccountName) string {
	acc := testConfiguration.Accounts[account]
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
		client_id       = "%s"
		client_secret   = "%s"

		experimental {
			embedded_client = %s
		}
	}
`, acc.Password, testConfiguration.ServerURL, acc.Email, acc.ClientID, acc.ClientSecret, testConfiguration.UseEmbeddedClientStr())
}

func tfConfigAttachmentSpecificPasswordManagerProvider() string {
	if testConfiguration.UseEmbeddedClient {
		return tfConfigPasswordManagerProvider(testAccountFullAdmin)
	}

	acc := testConfiguration.Accounts[testAccountFullAdmin]
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"

		experimental {
			embedded_client = false
		}
	}
`, acc.Password, testConfiguration.ReverseProxyServerURL, acc.Email)
}

func testOrRealSecretsManagerProvider(t *testing.T) (string, func()) {
	if IsOfficialBackend() {
		t.Logf("Using real Bitwarden Secrets Manager")
		return tfConfigSecretsManagerProvider(), func() {}
	}
	t.Logf("Spawning test Bitwarden Secrets Manager")
	return spawnTestSecretsManager(t)
}

func spawnTestSecretsManager(t *testing.T) (string, func()) {
	testSecretsManager := NewTestSecretsManager()
	ctx, stop := context.WithCancel(t.Context())
	go testSecretsManager.Run(ctx, 8081)

	orgId, err := testSecretsManager.ClientCreateNewOrganization()
	if err != nil {
		t.Fatal(err)
	}

	accessToken, err := testSecretsManager.ClientCreateAccessToken(orgId)
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"
		server = "http://localhost:8081"

		experimental {
			embedded_client = true
		}
	}
	`, accessToken), stop
}

func tfConfigSecretsManagerProvider() string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"
		server = "%s"
		experimental {
			embedded_client = true
		}
	}
`, os.Getenv("TEST_SECRETS_MANAGER_ACCESS_TOKEN"), testConfiguration.ServerURL)
}

func getObjectID(n string, objectId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		*objectId = rs.Primary.ID
		return nil
	}
}

func getAttachmentIDs(n string, objectId, itemId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		*objectId = rs.Primary.ID
		*itemId = rs.Primary.Attributes["item_id"]
		return nil
	}
}
