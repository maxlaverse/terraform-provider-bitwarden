package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
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
	_, err = webapiClient.LoginWithPassword(ctx, testEmail, testPassword, kdfIterations)
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
	t.Logf("Creating Folder")
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
	t.Log("Synced test client")

	areTestResourcesCreated = true
}

func ensureVaultwardenHasUser(t *testing.T) {
	userMu.Lock()
	defer userMu.Unlock()

	if isUserCreated {
		return
	}

	clearTestVault(t)

	webapiClient := webapi.NewClient(testServerURL)
	testUsername = fmt.Sprintf("test-%s", testUniqueIdentifier)
	testEmail = fmt.Sprintf("test-%s@laverse.net", testUniqueIdentifier)
	err := webapiClient.RegisterUser(context.Background(), testUsername, testEmail, testPassword, kdfIterations)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
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
	t.Log("Logged in test client")

	return client
}

func tfConfigProvider() string {
	useEmbeddedClient := "false"
	if os.Getenv("TEST_USE_EMBEDDED_CLIENT") == "1" {
		useEmbeddedClient = "true"
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
`, testPassword, testServerURL, testEmail, useEmbeddedClient)
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
