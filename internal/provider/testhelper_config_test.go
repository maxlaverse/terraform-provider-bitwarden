package provider

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

const (
	testDeviceIdentifer = "10a00887-3451-4607-8457-fcbfdc61faaa"
	testDeviceVersion   = "dev"
	testKdfIterations   = 5000
	defaultTestPassword = "test1234"

	accountTypePremium = "premium"
)

type testBackendType string

const (
	backendOfficial    testBackendType = "official"
	backendVaultwarden testBackendType = "vaultwarden"
)

type testAccountName string

const (
	testAccountFullAdmin  = "test-full-admin"
	testAccountOrgOwner   = "test-org-owner"
	testAccountOrgUser    = "test-org-user"
	testAccountOrgManager = "test-org-manager"
	testAccountOrgAdmin   = "test-org-admin"
)

type testAccount struct {
	Email                    string
	Name                     string
	Password                 string
	ClientID                 string
	ClientSecret             string
	AccountType              string
	UserIdInTestOrganization string
	RoleInTestOrganization   models.OrgMemberRoleType
}

type testConfigStruct struct {
	Accounts  map[testAccountName]testAccount
	Resources struct {
		OrganizationID string
		CollectionID   string
		FolderID       string
		GroupID        string
		GroupName      string
	}
	UniqueTestIdentifier          string
	ServerURL                     string
	ReverseProxyServerURL         string
	UseEmbeddedClient             bool
	Backend                       testBackendType
	wasAccountCreationAttempted   atomic.Bool
	didAccountCreationSucceed     bool
	wasResourcesCreationAttempted atomic.Bool
}

func (c *testConfigStruct) UseEmbeddedClientStr() string {
	if c.UseEmbeddedClient {
		return "true"
	}
	return "false"
}

func (c *testConfigStruct) ClientImplementationStr() string {
	if c.UseEmbeddedClient {
		return schema_definition.ClientImplementationEmbedded
	}
	return schema_definition.ClientImplementationCLI
}

func (c *testConfigStruct) WasResourcesCreationAttempted(t *testing.T) bool {
	if !c.wasResourcesCreationAttempted.CompareAndSwap(false, true) {
		return true
	}
	return false
}

func (c *testConfigStruct) WasAccountCreationAttempted(t *testing.T) bool {
	if IsOfficialBackend() {
		return true
	}

	if !c.wasAccountCreationAttempted.CompareAndSwap(false, true) {
		if !c.didAccountCreationSucceed {
			t.Fatal("Account creation failed")
		}
		return true
	}
	return false
}

var testConfiguration = testConfigStruct{
	Accounts: map[testAccountName]testAccount{},
}

func SkipIfOfficialBackend(t *testing.T, reason string) {
	if IsOfficialBackend() {
		t.Skipf("Skipping test as official backend is used: %s", reason)
	}
}

func SkipIfVaultwardenBackend(t *testing.T) {
	if testConfiguration.Backend == backendVaultwarden {
		t.Skip("Skipping test as vaultwarden backend is used")
	}
}

func SkipIfNonPremiumTestAccount(t *testing.T) {
	if testConfiguration.Accounts[testAccountFullAdmin].AccountType != accountTypePremium {
		t.Skip("Skipping test as non-premium test account is used")
	}
}

func SkipIfOfficialCLI(t *testing.T, reason string) {
	if !testConfiguration.UseEmbeddedClient {
		t.Skipf("Skipping test as official CLI is used: %s", reason)
	}
}

func IsOfficialBackend() bool {
	return testConfiguration.Backend == backendOfficial
}

func IsVaultwardenBackend() bool {
	return testConfiguration.Backend == backendVaultwarden
}

func init() {
	testConfiguration.Backend = testBackendType(os.Getenv("TEST_BACKEND"))
	if testConfiguration.Backend != backendOfficial && testConfiguration.Backend != backendVaultwarden {
		fmt.Printf("TEST_BACKEND must be either '%s' or '%s', got '%s'\n", backendOfficial, backendVaultwarden, testConfiguration.Backend)
		os.Exit(1)
	}

	loadEnvironmentVariablesFromFiles()
	loadTestServerConfiguration()
	loadTestAccountsConfiguration()
	loadTestResourcesConfiguration()
}

func loadEnvironmentVariablesFromFiles() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Error getting project root")
		os.Exit(1)
	}
	currentDir := filepath.Dir(file)
	projectRoot := path.Join(currentDir, "../../")
	_ = godotenv.Load(filepath.Join(projectRoot, ".env."+string(testConfiguration.Backend)))
}

func loadTestAccountsConfiguration() {
	if IsOfficialBackend() {
		loadOfficialBackendAccounts()
	} else {
		loadVaultwardenBackendAccounts()
	}
}

func loadOfficialBackendAccounts() {
	baseAccount := testAccount{
		Email:        os.Getenv("TEST_PASSWORD_MANAGER_BASE_EMAIL"),
		Password:     os.Getenv("TEST_PASSWORD_MANAGER_BASE_MASTER_PASSWORD"),
		ClientID:     os.Getenv("TEST_PASSWORD_MANAGER_BASE_CLIENT_ID"),
		ClientSecret: os.Getenv("TEST_PASSWORD_MANAGER_BASE_CLIENT_SECRET"),
		AccountType:  os.Getenv("TEST_PASSWORD_MANAGER_BASE_ACCOUNT_TYPE"),
		Name:         os.Getenv("TEST_PASSWORD_MANAGER_BASE_NAME"),
	}

	testConfiguration.Accounts[testAccountFullAdmin] = baseAccount
	testConfiguration.Accounts[testAccountOrgOwner] = testAccount{
		Email:                    os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_EMAIL"),
		Password:                 os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_MASTER_PASSWORD"),
		ClientID:                 os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_CLIENT_ID"),
		ClientSecret:             os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_CLIENT_SECRET"),
		AccountType:              os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_ACCOUNT_TYPE"),
		Name:                     os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_NAME"),
		UserIdInTestOrganization: os.Getenv("TEST_PASSWORD_MANAGER_ORG_OWNER_ORGANIZATION_MEMBER_ID"),
		RoleInTestOrganization:   models.OrgMemberRoleTypeOwner,
	}
	testConfiguration.Accounts[testAccountOrgUser] = testAccount{
		Email:                    os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_EMAIL"),
		Password:                 os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_MASTER_PASSWORD"),
		ClientID:                 os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_CLIENT_ID"),
		ClientSecret:             os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_CLIENT_SECRET"),
		AccountType:              os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_ACCOUNT_TYPE"),
		Name:                     os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_NAME"),
		UserIdInTestOrganization: os.Getenv("TEST_PASSWORD_MANAGER_ORG_USER_ORGANIZATION_MEMBER_ID"),
		RoleInTestOrganization:   models.OrgMemberRoleTypeUser,
	}
}

func loadVaultwardenBackendAccounts() {
	testConfiguration.UniqueTestIdentifier = fmt.Sprintf("%02d%02d%02d", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	baseName := fmt.Sprintf("test-%s", testConfiguration.UniqueTestIdentifier)
	basePassword := defaultTestPassword

	accounts := map[testAccountName]testAccount{
		testAccountFullAdmin: {
			Name:        baseName,
			Email:       fmt.Sprintf("%s@laverse.net", baseName),
			Password:    basePassword,
			AccountType: accountTypePremium,
		},
		testAccountOrgOwner: {
			Name:                   baseName,
			Email:                  fmt.Sprintf("%s-org-owner@laverse.net", baseName),
			Password:               basePassword,
			AccountType:            accountTypePremium,
			RoleInTestOrganization: models.OrgMemberRoleTypeOwner,
		},
		testAccountOrgUser: {
			Name:                   baseName,
			Email:                  fmt.Sprintf("%s-org-user@laverse.net", baseName),
			Password:               basePassword,
			AccountType:            accountTypePremium,
			RoleInTestOrganization: models.OrgMemberRoleTypeUser,
		},
		testAccountOrgAdmin: {
			Name:                   baseName,
			Email:                  fmt.Sprintf("%s-org-admin@laverse.net", baseName),
			Password:               basePassword,
			AccountType:            accountTypePremium,
			RoleInTestOrganization: models.OrgMemberRoleTypeAdmin,
		},
		testAccountOrgManager: {
			Name:                   baseName,
			Email:                  fmt.Sprintf("%s-org-manager@laverse.net", baseName),
			Password:               basePassword,
			AccountType:            accountTypePremium,
			RoleInTestOrganization: models.OrgMemberRoleTypeManager,
		},
	}

	testConfiguration.Accounts = accounts
}

func loadTestServerConfiguration() {
	testConfiguration.UseEmbeddedClient = os.Getenv("TEST_EXPERIMENTAL_EMBEDDED_CLIENT") == "1"
	testConfiguration.ServerURL = os.Getenv("TEST_SERVER_URL")
	testConfiguration.ReverseProxyServerURL = os.Getenv("TEST_REVERSE_PROXY_URL")
	if testConfiguration.ReverseProxyServerURL == "" {
		testConfiguration.ReverseProxyServerURL = testConfiguration.ServerURL
	}
}

func loadTestResourcesConfiguration() {
	if !IsOfficialBackend() {
		return
	}

	testConfiguration.Resources.CollectionID = os.Getenv("TEST_PASSWORD_MANAGER_COLLECTION_ID")
	testConfiguration.Resources.FolderID = os.Getenv("TEST_PASSWORD_MANAGER_FOLDER_ID")
	testConfiguration.Resources.OrganizationID = os.Getenv("TEST_PASSWORD_MANAGER_ORGANIZATION_ID")
}

func (c *testConfigStruct) PrintConfiguration() string {
	var buf strings.Builder

	buf.WriteString("\nTest Configuration:\n")
	buf.WriteString("------------------\n")
	fmt.Fprintf(&buf, "Backend: %s\n", c.Backend)
	fmt.Fprintf(&buf, "Server URL: %s\n", c.ServerURL)
	fmt.Fprintf(&buf, "Reverse Proxy URL: %s\n", c.ReverseProxyServerURL)
	fmt.Fprintf(&buf, "Use Embedded Client: %v\n", c.UseEmbeddedClient)
	fmt.Fprintf(&buf, "Unique Test Identifier: %s\n", c.UniqueTestIdentifier)

	buf.WriteString("\nResources:\n")
	fmt.Fprintf(&buf, "  Organization ID: %s\n", maskIdentifier(c.Resources.OrganizationID))
	fmt.Fprintf(&buf, "  Collection ID: %s\n", maskIdentifier(c.Resources.CollectionID))
	fmt.Fprintf(&buf, "  Folder ID: %s\n", maskIdentifier(c.Resources.FolderID))
	fmt.Fprintf(&buf, "  Group ID: %s\n", maskIdentifier(c.Resources.GroupID))
	fmt.Fprintf(&buf, "  Group Name: %s\n", maskIfEmpty(c.Resources.GroupName))

	for name, account := range c.Accounts {
		fmt.Fprintf(&buf, "\n Account '%s':\n", name)
		fmt.Fprintf(&buf, "  Name: %s\n", maskIfEmpty(account.Name))
		fmt.Fprintf(&buf, "  Email: %s\n", maskEmail(account.Email))
		fmt.Fprintf(&buf, "  Password: %s\n", maskSensitive(account.Password))
		fmt.Fprintf(&buf, "  Client ID: %s\n", maskSensitive(account.ClientID))
		fmt.Fprintf(&buf, "  Client Secret: %s\n", maskSensitive(account.ClientSecret))
		fmt.Fprintf(&buf, "  Account Type: %s\n", maskIfEmpty(account.AccountType))
		fmt.Fprintf(&buf, "  Organization Role: %s\n", maskIfEmpty(account.RoleInTestOrganization.String()))
		fmt.Fprintf(&buf, "  Organization Member ID: %s\n", maskIdentifier(account.UserIdInTestOrganization))
	}
	buf.WriteString("------------------\n")

	return buf.String()
}

func maskIfEmpty(value string) string {
	if value == "" {
		return "<not set>"
	}
	return value
}

func maskIdentifier(value string) string {
	if value == "" {
		return "<not set>"
	}

	parts := strings.Split(value, "-")
	if len(parts) != 5 {
		return value
	}

	for i := 0; i < len(parts)-1; i++ {
		parts[i] = strings.Repeat("*", len(parts[i]))
	}

	return strings.Join(parts, "-")
}

func maskEmail(value string) string {
	if value == "" {
		return "<not set>"
	}
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return value
	}
	parts[0] = strings.Repeat("*", len(parts[0]))
	return strings.Join(parts, "@")
}

func maskSensitive(value string) string {
	if value == "" {
		return "<not set>"
	}
	return strings.Repeat("*", len(value))
}
