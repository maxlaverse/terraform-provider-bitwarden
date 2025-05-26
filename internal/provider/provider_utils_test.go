package provider

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/joho/godotenv"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	// Constants used to interact with a test Vaultwarden instance
	testDeviceIdentifer = "10a00887-3451-4607-8457-fcbfdc61faaa"
	testDeviceVersion   = "dev"
	testKdfIterations   = 5000

	// Backend types
	backendOfficial    = "official"
	backendVaultwarden = "vaultwarden"

	accountTypePremium = "premium"
)

// Generated resources used for testing
var testEmail string
var testBackend string
var testMasterPassword = "test1234"
var testClientID string
var testClientSecret string
var testAccountType = accountTypePremium
var testAccountNameOrgOwner string
var testAccountEmailOrgOwner string
var testAccountEmailOrgUser string
var testAccountEmailOrgAdmin string
var testAccountEmailOrgManager string

var testAccountEmailOrgOwnerInTestOrgUserId string
var testAccountEmailOrgUserInTestOrgUserId string
var testAccountEmailOrgAdminInTestOrgUserId string
var testAccountEmailOrgManagerInTestOrgUserId string

var testUsername string
var testServerURL = "http://127.0.0.1:8000/"
var testReverseProxyServerURL string
var testOrganizationID string
var testCollectionID string
var testFolderID string
var testGroupID string
var testGroupName string
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
	testBackend = os.Getenv("TEST_BACKEND")
	if testBackend == "" {
		fmt.Println("TEST_BACKEND environment variable is not set")
		os.Exit(1)
	}

	if testBackend != backendOfficial && testBackend != backendVaultwarden {
		fmt.Printf("TEST_BACKEND must be either '%s' or '%s', got '%s'\n", backendOfficial, backendVaultwarden, testBackend)
		os.Exit(1)
	}

	// Get the project root directory
	projectRoot, err := getProjectRoot()
	if err != nil {
		fmt.Printf("Error getting project root: %v\n", err)
		os.Exit(1)
	}

	_ = godotenv.Load(filepath.Join(projectRoot, ".env."+testBackend))
	_ = godotenv.Load(filepath.Join(projectRoot, ".env"))

	testUniqueIdentifier = fmt.Sprintf("%02d%02d%02d", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	testAccountNameOrgOwner = fmt.Sprintf("test-%s", testUniqueIdentifier)

	// Load environment variables
	loadEnvironmentVariables()
}

// getProjectRoot returns the absolute path to the project root directory
func getProjectRoot() (string, error) {
	// Start from the current directory (where the test is running)
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to get caller information")
	}
	currentDir := filepath.Dir(file)
	return path.Join(currentDir, "../../"), nil
}

// loadEnvironmentVariables loads all environment variables used in tests
func loadEnvironmentVariables() {
	if os.Getenv("TEST_EXPERIMENTAL_EMBEDDED_CLIENT") == "1" {
		useEmbeddedClient = true
	}

	if v := os.Getenv("TEST_PASSWORD_MANAGER_MASTER_PASSWORD"); v != "" {
		testMasterPassword = v
	}

	if v := os.Getenv("TEST_PASSWORD_MANAGER_EMAIL"); v != "" {
		testEmail = v
	}

	if v := os.Getenv("TEST_PASSWORD_MANAGER_CLIENT_ID"); v != "" {
		testClientID = v
	}

	if v := os.Getenv("TEST_PASSWORD_MANAGER_CLIENT_SECRET"); v != "" {
		testClientSecret = v
	}

	if v := os.Getenv("TEST_PASSWORD_MANAGER_ACCOUNT_TYPE"); v != "" {
		testAccountType = v
	}

	if v := os.Getenv("TEST_SERVER_URL"); v != "" {
		testServerURL = v
	}

	if v := os.Getenv("TEST_REVERSE_PROXY_URL"); v != "" {
		testReverseProxyServerURL = v
	} else {
		testReverseProxyServerURL = testServerURL
	}

	// When using the official backend, we reuse existing resources rather than creating new ones
	// to avoid hitting free account limits on organizations and collections.
	if IsOfficialBackend() {
		testAccountEmailOrgOwner = testEmail

		if v := os.Getenv("TEST_PASSWORD_MANAGER_COLLECTION_ID"); v != "" {
			testCollectionID = v
		}
		if v := os.Getenv("TEST_PASSWORD_MANAGER_FOLDER_ID"); v != "" {
			testFolderID = v
		}
		if v := os.Getenv("TEST_PASSWORD_MANAGER_ORGANIZATION_ID"); v != "" {
			testOrganizationID = v
		}
		if v := os.Getenv("TEST_PASSWORD_MANAGER_ORGANIZATION_MEMBER_ID"); v != "" {
			testAccountEmailOrgOwnerInTestOrgUserId = v
		}

		if v := os.Getenv("TEST_PASSWORD_MANAGER_ORGANIZATION_OTHER_MEMBER_ID"); v != "" {
			testAccountEmailOrgUserInTestOrgUserId = v
		}

		if v := os.Getenv("TEST_PASSWORD_MANAGER_USER_NAME"); v != "" {
			testAccountNameOrgOwner = v
		}
	}
}

func SkipIfOfficialBackend(t *testing.T, reason string) {
	if IsOfficialBackend() {
		t.Skipf("Skipping test as official backend is used: %s", reason)
	}
}

func SkipIfVaultwardenBackend(t *testing.T) {
	if testBackend == backendVaultwarden {
		t.Skip("Skipping test as vaultwarden backend is used")
	}
}

func SkipIfNonPremiumTestAccount(t *testing.T) {
	if testAccountType != accountTypePremium {
		t.Skip("Skipping test as non-premium test account is used")
	}
}

func SkipIfOfficialCLI(t *testing.T, reason string) {
	if !useEmbeddedClient {
		t.Skipf("Skipping test as official CLI is used: %s", reason)
	}
}

func IsOfficialBackend() bool {
	return testBackend == backendOfficial
}

func IsVaultwardenBackend() bool {
	return testBackend == backendVaultwarden
}

func ensureVaultwardenConfigured(t *testing.T) {
	testResourcesMu.Lock()
	defer testResourcesMu.Unlock()

	if areTestResourcesCreated || IsOfficialBackend() {
		return
	}

	// Force refreshing the embedded client variable
	tfConfigPasswordManagerProvider()

	ensureVaultwardenHasUser(t)
	ctx := context.Background()

	bwClient := bwTestClient(t)

	var err error
	testOrganizationID, err = bwClient.(embedded.PasswordManagerClient).CreateOrganization(ctx, "org-"+testUniqueIdentifier, "coll-"+testUniqueIdentifier, testEmail)
	if err != nil {
		t.Fatal(err)
	}

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgOwner, models.OrgMemberRoleTypeOwner)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgOwnerInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgOwner)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Invited %s to organization %s (%s)", testAccountEmailOrgOwner, testOrganizationID, testAccountEmailOrgOwnerInTestOrgUserId)

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgAdmin, models.OrgMemberRoleTypeAdmin)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgAdminInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgAdmin)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Invited %s to organization %s (%s)", testAccountEmailOrgAdmin, testOrganizationID, testAccountEmailOrgAdminInTestOrgUserId)

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgUser, models.OrgMemberRoleTypeUser)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgUserInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgUser)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Invited %s to organization %s (%s)", testAccountEmailOrgUser, testOrganizationID, testAccountEmailOrgUserInTestOrgUserId)

	err = bwClient.(embedded.PasswordManagerClient).InviteUser(ctx, testOrganizationID, testAccountEmailOrgManager, models.OrgMemberRoleTypeManager)
	if err != nil {
		t.Fatal(err)
	}

	testAccountEmailOrgManagerInTestOrgUserId, err = bwClient.(embedded.PasswordManagerClient).ConfirmInvite(ctx, testOrganizationID, testAccountEmailOrgManager)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Invited %s to organization %s (%s)", testAccountEmailOrgManager, testOrganizationID, testAccountEmailOrgManagerInTestOrgUserId)

	webapiClient := webapi.NewClient(testServerURL, embedded.NewDeviceIdentifier(), testDeviceVersion)
	_, err = webapiClient.LoginWithPassword(ctx, testEmail, testMasterPassword, models.KdfConfiguration{KdfIterations: testKdfIterations})
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

	testGroupName = fmt.Sprintf("group-%s-bar", testUniqueIdentifier)
	group, err := bwClient.CreateOrganizationGroup(ctx, models.OrgGroup{
		OrganizationID: testOrganizationID,
		Name:           testGroupName,
		Collections:    []models.OrgCollectionMember{},
		Users:          []models.OrgCollectionMember{},
	})
	if err != nil {
		t.Fatal(err)
	}

	testGroupID = group.ID
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Created Group '%s' (%s)", testGroupName, testGroupID)

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

	if isUserCreated || IsOfficialBackend() {
		return
	}

	clearTestVault(t)

	client := embedded.NewPasswordManagerClient(testServerURL, testDeviceIdentifer, testDeviceVersion)
	testUsername = fmt.Sprintf("test-%s", testUniqueIdentifier)
	testEmail = fmt.Sprintf("test-%s@laverse.net", testUniqueIdentifier)
	kdfConfig := models.KdfConfiguration{
		KdfType:        models.KdfTypePBKDF2_SHA256,
		KdfIterations:  testKdfIterations,
		KdfMemory:      0,
		KdfParallelism: 0,
	}
	err := client.RegisterUser(context.Background(), testUsername, testEmail, testMasterPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Created test user (main) %s", testEmail)

	testAccountEmailOrgOwner = fmt.Sprintf("test-%s-org-owner@laverse.net", testUniqueIdentifier)
	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgOwner, testMasterPassword, kdfConfig)
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
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgUser, testMasterPassword, kdfConfig)
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
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgAdmin, testMasterPassword, kdfConfig)
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
	err = client.RegisterUser(context.Background(), testUsername, testAccountEmailOrgManager, testMasterPassword, kdfConfig)
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
	var err error
	if testClientID != "" {
		err = client.LoginWithAPIKey(context.Background(), testMasterPassword, testClientID, testClientSecret)
	} else {
		err = client.LoginWithPassword(context.Background(), testEmail, testMasterPassword)
	}
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

	client := bwcli.NewPasswordManagerClient(bwcli.DisableRetryBackoff(), bwcli.WithAppDataDir(vault))
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
			if testClientID != "" {
				err = client.LoginWithAPIKey(context.Background(), testMasterPassword, testClientID, testClientSecret)
			} else {
				err = client.LoginWithPassword(context.Background(), testEmail, testMasterPassword)
			}

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
		err = client.Unlock(context.Background(), testMasterPassword)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Logf("Official test client already logged-in: %s", status.Status)
	}
	return client
}

func tfConfigPasswordManagerProvider() string {
	useEmbeddedClientStr := "false"
	if useEmbeddedClient {
		useEmbeddedClientStr = "true"
	}

	if len(testClientID) > 0 && len(testClientSecret) > 0 {
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
`, testMasterPassword, testServerURL, testEmail, testClientID, testClientSecret, useEmbeddedClientStr)
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
`, testMasterPassword, testServerURL, testEmail, useEmbeddedClientStr)
}

func testOrRealSecretsManagerProvider(t *testing.T) (string, func()) {
	tfProvider := tfConfigSecretsManagerProvider()
	if IsOfficialBackend() {
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

func tfConfigSecretsManagerProvider() string {
	accessToken := os.Getenv("TEST_SECRETS_MANAGER_ACCESS_TOKEN")
	return fmt.Sprintf(`
	provider "bitwarden" {
		access_token = "%s"
		server = "%s"
		experimental {
			embedded_client = true
		}
	}
`, accessToken, testServerURL)
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
