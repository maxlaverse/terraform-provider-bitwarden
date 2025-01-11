package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	// Constants used to interact with a test Vaultwarden instance
	testPassword        = "test1234"
	testDeviceIdentifer = "10a00887-3451-4607-8457-fcbfdc61faaa"
	testDeviceVersion   = "dev"
	kdfIterations       = 10000
)

// Generated resources used for testing
var testEmail string
var testAccountEmailOrgOwner string
var testAccountEmailOrgUser string
var testAccountEmailOrgAdmin string
var testAccountEmailOrgManager string

var testAccountEmailOrgOwnerInTestOrgUserId string
var testAccountEmailOrgUserInTestOrgUserId string
var testAccountEmailOrgAdminInTestOrgUserId string
var testAccountEmailOrgManagerInTestOrgUserId string

var testUsername string
var testServerURL string
var testOrganizationID string
var testCollectionID string
var testFolderID string
var testUniqueIdentifier string
var useEmbeddedClient bool

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"bitwarden": func() (*schema.Provider, error) {
		return New(versionTestDisabledRetries)(), nil
	},
}

var areTestResourcesCreated bool
var testResourcesMu sync.Mutex
var isUserCreated bool
var userMu sync.Mutex

func init() {
	host := os.Getenv("VAULTWARDEN_HOST")
	port := os.Getenv("VAULTWARDEN_PORT")

	if len(host) == 0 {
		host = "127.0.0.1"
	}
	if len(port) == 0 {
		port = "8080"
	}

	testServerURL = fmt.Sprintf("http://%s:%s/", host, port)
	testUniqueIdentifier = fmt.Sprintf("%02d%02d%02d", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
}

func ensureVaultwardenConfigured(t *testing.T) {
	testResourcesMu.Lock()
	defer testResourcesMu.Unlock()

	if areTestResourcesCreated {
		return
	}

	ensureVaultwardenHasUser(t)
	ctx := context.Background()

	bwClient := bwTestClient(t)

	var err error
	testOrganizationID, err = bwClient.(embedded.PasswordManagerClient).CreateOrganization(ctx, "org-"+testUniqueIdentifier, "coll-"+testUniqueIdentifier, testEmail)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Inviting users to organization %s", testOrganizationID)
	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgOwner, models.OrgMemberRoleTypeOwner)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgOwnerInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgOwner)
	if err != nil {
		t.Fatal(err)
	}

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgAdmin, models.OrgMemberRoleTypeAdmin)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgAdminInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgAdmin)
	if err != nil {
		t.Fatal(err)
	}

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgUser, models.OrgMemberRoleTypeUser)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgUserInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgUser)
	if err != nil {
		t.Fatal(err)
	}

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgManager, models.OrgMemberRoleTypeManager)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgManagerInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgManager)
	if err != nil {
		t.Fatal(err)
	}

	webapiClient := webapi.NewClient(testServerURL, embedded.NewDeviceIdentifier(), testDeviceVersion)
	_, err = webapiClient.LoginWithPassword(ctx, testEmail, testPassword, models.KdfConfiguration{KdfIterations: kdfIterations})
	if err != nil {
		t.Fatal(err)
	}
	cols, err := webapiClient.GetOrganizationCollections(ctx, testOrganizationID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) == 0 {
		t.Fatal("No collections found")
	}
	testCollectionID = cols[0].Id

	testFolderName := fmt.Sprintf("folder-%s-bar", testUniqueIdentifier)
	folder, err := bwClient.CreateFolder(ctx, models.Folder{
		Object: models.ObjectTypeFolder,
		Name:   testFolderName,
	})
	if err != nil {
		t.Fatal(err)
	}

	testFolderID = folder.ID
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Created Folder '%s' (%s)", testFolderName, testFolderID)

	err = bwClient.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Synced embedded test client")

	areTestResourcesCreated = true
}

func ensureVaultwardenHasUser(t *testing.T) {
	userMu.Lock()
	defer userMu.Unlock()

	if isUserCreated {
		return
	}

	clearTestVault(t)

	client := embedded.NewPasswordManagerClient(testServerURL, testDeviceIdentifer, testDeviceVersion)
	testUsername = fmt.Sprintf("test-%s", testUniqueIdentifier)
	testEmail = fmt.Sprintf("test-%s@laverse.net", testUniqueIdentifier)
	kdfConfig := models.KdfConfiguration{
		KdfType:        models.KdfTypePBKDF2_SHA256,
		KdfIterations:  kdfIterations,
		KdfMemory:      0,
		KdfParallelism: 0,
	}
	err := client.RegisterUser(context.Background(), testUsername, testEmail, testPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}

	testAccountEmailOrgOwner = fmt.Sprintf("test-%s-org-owner@laverse.net", testUniqueIdentifier)
	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgOwner, testPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Created test user (org-owner) %s", testAccountEmailOrgOwner)

	testAccountEmailOrgUser = fmt.Sprintf("test-%s-org-user@laverse.net", testUniqueIdentifier)
	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgUser, testPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Created test user (org-user) %s", testAccountEmailOrgUser)

	testAccountEmailOrgAdmin = fmt.Sprintf("test-%s-org-admin@laverse.net", testUniqueIdentifier)
	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgAdmin, testPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Created test user (org-admin) %s", testAccountEmailOrgAdmin)

	testAccountEmailOrgManager = fmt.Sprintf("test-%s-org-manager@laverse.net", testUniqueIdentifier)
	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgManager, testPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Created test user (org-manager) %s", testAccountEmailOrgManager)

	isUserCreated = true
}

func clearTestVault(t *testing.T) {
	err := os.Remove(".bitwarden/data.json")
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func bwTestClient(t *testing.T) bitwarden.PasswordManager {
	client := embedded.NewPasswordManagerClient(testServerURL, testDeviceIdentifer, testDeviceVersion)
	err := client.LoginWithPassword(context.Background(), testEmail, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Logged in embedded test client")

	return client
}

func bwOfficialTestClient(t *testing.T) bwcli.PasswordManagerClient {
	vault, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	bwExec, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	client := bwcli.NewPasswordManagerClient(bwExec, bwcli.DisableRetryBackoff(), bwcli.WithAppDataDir(vault))
	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ServerURL) == 0 {
		err = client.SetServer(context.Background(), testServerURL)
		if err != nil {
			t.Fatal(err)
		}
	}
	if status.Status == bwcli.StatusUnauthenticated {

		retries := 0
		for {
			err = client.LoginWithPassword(context.Background(), testEmail, testPassword)
			if err != nil {
				// Retry if the user creation hasn't been fully taken into account yet
				if retries < 3 {
					retries++
					t.Log("Account creation not taken into account yet, retrying...")
					time.Sleep(time.Duration(retries) * time.Second)
					continue
				}
				t.Fatal(err)
			}
			break
		}
	} else if status.Status == bwcli.StatusLocked {
		err = client.Unlock(context.Background(), testPassword)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Logf("Official test client already logged-in: %s", status.Status)
	}
	return client
}

func tfConfigPasswordManagerProvider() string {
	if os.Getenv("TEST_USE_EMBEDDED_CLIENT") == "1" {
		useEmbeddedClient = true
	}

	useEmbeddedClientStr := "false"
	if useEmbeddedClient {
		useEmbeddedClientStr = "true"
	}
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"

		experimental {
			embedded_client = %s
		}
	}
`, testPassword, testServerURL, testEmail, useEmbeddedClientStr)
}

func testOrRealSecretsManagerProvider(t *testing.T) (string, func()) {
	tfProvider, defined := tfConfigSecretsManagerProvider()
	if defined {
		t.Logf("Using real Bitwarden Secrets Manager")
		stop := func() {}
		return tfProvider, stop
	} else {
		t.Logf("Spawning test Bitwarden Secrets Manager")
		return spawnTestSecretsManager(t)
	}
}

func spawnTestSecretsManager(t *testing.T) (string, func()) {
	testSecretsManager := NewTestSecretsManager()
	ctx, stop := context.WithCancel(context.Background())
	go testSecretsManager.Run(ctx, 8081)

	orgId, err := testSecretsManager.ClientCreateNewOrganization()
	if err != nil {
		t.Fatal(err)
	}

	accessToken, err := testSecretsManager.ClientCreateAccessToken(orgId)
	if err != nil {
		t.Fatal(err)
	}

	providerConfiguration := fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"
		server = "http://localhost:8081"

		experimental {
			embedded_client = true
		}
	}
		`, accessToken)
	return providerConfiguration, stop
}

func tfConfigSecretsManagerProvider() (string, bool) {
	accessToken := os.Getenv("TEST_REAL_BWS_ACCESS_TOKEN")
	return fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"

		experimental {
			embedded_client = true
		}
	}
`, accessToken), len(accessToken) > 0
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
