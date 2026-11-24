package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

// cliSession is the BW_SESSION for the account currently logged into the
// shared test vault directory. CLI acceptance tests reuse it so Configure
// skips Unlock (and its KDF) on every Terraform command.
type cliSession struct {
	account testAccountName
	key     string
}

var (
	cliSessionMu     sync.Mutex
	cachedCLISession cliSession
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bitwarden": func() (tfprotov6.ProviderServer, error) {
		opts := []ProviderOption{}
		if !IsOfficialBackend() {
			opts = append(opts, WithDisableRetries())
		}
		if !testConfiguration.UseEmbeddedClient {
			opts = append(opts, WithDisableConfigureSync())
		}
		return NewProviderServer("", opts...)(), nil
	},
}

func ensureTestConfigurationReady(t *testing.T) {
	bindTestName(t)
	ensureTestAccountsExist(t)
	ensureTestResourcesExist(t)
	ensureCLISession(t)
}

func ensureCLISession(t *testing.T) {
	t.Helper()
	if testConfiguration.UseEmbeddedClient {
		return
	}
	_ = bwCLITestClient(t)
}

func rememberCLISession(account testAccountName, key string) {
	cliSessionMu.Lock()
	defer cliSessionMu.Unlock()
	cachedCLISession = cliSession{account: account, key: key}
}

func cliSessionKeyFor(account testAccountName) (string, bool) {
	if testConfiguration.UseEmbeddedClient {
		return "", false
	}
	cliSessionMu.Lock()
	defer cliSessionMu.Unlock()
	if cachedCLISession.key == "" || cachedCLISession.account != account {
		return "", false
	}
	return cachedCLISession.key, true
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
	if key, ok := cliSessionKeyFor(testAccountFullAdmin); ok {
		client.SetSessionKey(key)
	}
	status, err := client.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	acc := testConfiguration.Accounts[testAccountFullAdmin]
	loggedIn := status.Status == bwcli.StatusLocked || status.Status == bwcli.StatusUnlocked
	wrongIdentity := loggedIn && (!status.VaultFromServer(testConfiguration.ServerURL) || (status.UserEmail != "" && !status.VaultOfUser(acc.Email)))
	if wrongIdentity {
		rememberCLISession("", "")
		client.SetSessionKey("")
		if err := client.Logout(t.Context()); err != nil {
			t.Fatal(err)
		}
		status.Status = bwcli.StatusUnauthenticated
	}

	if !status.VaultFromServer(testConfiguration.ServerURL) {
		if err := client.SetServer(t.Context(), testConfiguration.ServerURL); err != nil {
			t.Fatal(err)
		}
	}

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
	if key := client.GetSessionKey(); key != "" {
		rememberCLISession(testAccountFullAdmin, key)
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
	if key, ok := cliSessionKeyFor(account); ok {
		return fmt.Sprintf(`
	provider "bitwarden" {
		server          = "%s"
		email           = "%s"
		session_key     = "%s"
		client_implementation = "%s"
	}
`, testConfiguration.ServerURL, acc.Email, key, testConfiguration.ClientImplementationStr())
	}
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
		client_id       = "%s"
		client_secret   = "%s"
		client_implementation = "%s"
	}
`, acc.Password, testConfiguration.ServerURL, acc.Email, acc.ClientID, acc.ClientSecret, testConfiguration.ClientImplementationStr())
}

func tfConfigAttachmentSpecificPasswordManagerProvider() string {
	if testConfiguration.UseEmbeddedClient {
		return tfConfigPasswordManagerProvider(testAccountFullAdmin)
	}

	// session_key Configure compares server URLs; a reverse proxy that is not
	// the vault's configured server would log out and then fail to re-login.
	if testConfiguration.ReverseProxyServerURL == "" || testConfiguration.ReverseProxyServerURL == testConfiguration.ServerURL {
		return tfConfigPasswordManagerProvider(testAccountFullAdmin)
	}

	acc := testConfiguration.Accounts[testAccountFullAdmin]
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
		client_implementation = "%s"
	}
`, acc.Password, testConfiguration.ReverseProxyServerURL, acc.Email, testConfiguration.ClientImplementationStr())
}

func testOrRealSecretsManagerProvider(t *testing.T) (string, func()) {
	bindTestName(t)
	if IsOfficialBackend() {
		if os.Getenv("TEST_SECRETS_MANAGER_ACCESS_TOKEN") == "" {
			t.Fatal("TEST_BACKEND=official requires TEST_SECRETS_MANAGER_ACCESS_TOKEN")
		}
		if isLocalTestServerURL(testConfiguration.ServerURL) {
			t.Fatalf("TEST_BACKEND=official must not use a local Secrets Manager, got %q", testConfiguration.ServerURL)
		}
		t.Logf("Using real Bitwarden Secrets Manager")
		return tfConfigSecretsManagerProvider(), func() {}
	}
	t.Logf("Spawning test Bitwarden Secrets Manager")
	return spawnTestSecretsManager(t)
}

func spawnTestSecretsManager(t *testing.T) (string, func()) {
	if IsOfficialBackend() {
		t.Fatal("refusing to spawn mock Secrets Manager when TEST_BACKEND=official")
	}

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
	t.Logf("Created access token: %s", accessToken)

	return fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"
		server = "http://localhost:8081"
		client_implementation = "%s"
	}
	`, accessToken, testConfiguration.ClientImplementationStr()), stop
}

func tfConfigSecretsManagerProvider() string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"
		server = "%s"
		client_implementation = "%s"
	}
`, os.Getenv("TEST_SECRETS_MANAGER_ACCESS_TOKEN"), testConfiguration.ServerURL, testConfiguration.ClientImplementationStr())
}

// updateItemOutsideTerraform applies a mutation as if it came from the Bitwarden
// UI or another client, then syncs the CLI vault when tests are not using the
// embedded client.
func updateItemOutsideTerraform(t *testing.T, objectID string, mutate func(*models.Item)) {
	t.Helper()

	client := bwEmbeddedTestClient(t)
	item, err := client.GetItem(t.Context(), models.Item{ID: objectID, Object: models.ObjectTypeItem})
	if err != nil {
		t.Fatalf("getting item %s: %v", objectID, err)
	}
	mutate(item)
	if _, err := client.EditItem(t.Context(), *item); err != nil {
		t.Fatalf("editing item %s: %v", objectID, err)
	}
	if testConfiguration.UseEmbeddedClient {
		return
	}
	if err := bwCLITestClient(t).Sync(t.Context()); err != nil {
		t.Fatalf("syncing CLI vault after out-of-band edit: %v", err)
	}
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
