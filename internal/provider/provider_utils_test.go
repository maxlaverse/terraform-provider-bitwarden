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
	testPassword  = "test1234"
	kdfIterations = 10000
)

// Generated resources used for testing
var testEmail string
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
		return New(versionDev)(), nil
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
	testOrganizationID, err = bwClient.(embedded.WebAPIVault).CreateOrganization(ctx, "org-"+testUniqueIdentifier, "coll-"+testUniqueIdentifier, testEmail)
	if err != nil {
		t.Fatal(err)
	}

	webapiClient := webapi.NewClient(testServerURL)
	_, err = webapiClient.LoginWithPassword(ctx, testEmail, testPassword, models.KdfConfiguration{KdfIterations: kdfIterations})
	if err != nil {
		t.Fatal(err)
	}
	cols, err := webapiClient.GetCollections(ctx, testOrganizationID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) == 0 {
		t.Fatal("No collections found")
	}
	testCollectionID = cols[0].Id

	testFolderName := fmt.Sprintf("folder-%s-bar", testUniqueIdentifier)
	folder, err := bwClient.CreateObject(ctx, models.Object{
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

	client := embedded.NewWebAPIVault(testServerURL)
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

	testEmail2 := fmt.Sprintf("test-%s-argon2@laverse.net", testUniqueIdentifier)
	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}
	err = client.RegisterUser(context.Background(), testUsername, testEmail2, testPassword, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Created test user (argon2) %s", testEmail2)
	isUserCreated = true
}

func clearTestVault(t *testing.T) {
	err := os.Remove(".bitwarden/data.json")
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func bwTestClient(t *testing.T) bitwarden.Client {
	client := embedded.NewWebAPIVault(testServerURL)
	err := client.LoginWithPassword(context.Background(), testEmail, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Logged in embedded test client")

	return client
}

func bwOfficialTestClient(t *testing.T) bwcli.CLIClient {
	vault, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	bwExec, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	client := bwcli.NewClient(bwExec, bwcli.DisableRetryBackoff(), bwcli.WithAppDataDir(vault))
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

func tfConfigProvider() string {
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
