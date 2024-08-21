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
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	// Constants used to interact with a test Vaultwarden instance
	testEmail     = "test@laverse.net"
	testPassword  = "test1234"
	kdfIterations = 10000
)

// Generated resources used for testing
var testServerURL string
var testOrganizationID string
var testCollectionID string
var testFolderName string
var testFolderID string
var testSessionKey string

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
}

func ensureVaultwardenHasUser(t *testing.T) {
	userMu.Lock()
	defer userMu.Unlock()

	if isUserCreated {
		return
	}

	webapiClient := webapi.NewClient(testServerURL)

	err := webapiClient.RegisterUser("test", testEmail, testPassword, kdfIterations)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	isUserCreated = true

	bwTestClient(t)
}

func ensureVaultwardenConfigured(t *testing.T) {
	start := time.Now()
	testResourcesMu.Lock()
	defer func() {
		testResourcesMu.Unlock()
		t.Logf("ensureVaultwardenConfigured() took %s", time.Since(start))
	}()

	if areTestResourcesCreated {
		return
	}

	webapiClient := webapi.NewClient(testServerURL)

	userAlreadyExists := false
	err := webapiClient.RegisterUser("test", testEmail, testPassword, kdfIterations)
	if err != nil && strings.Contains(err.Error(), "User already exists") {
		userAlreadyExists = true
	}

	err = webapiClient.Login(testEmail, testPassword, kdfIterations)
	if err != nil {
		if userAlreadyExists {
			t.Fatalf("Unable to log into test instance, and the user was already present. Try removing it! Error: %v", err)
		} else {
			t.Fatal(err)
		}
	}
	dateTimeStr := fmt.Sprintf("%02d%02d%02d", time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	testOrganizationID, err = webapiClient.CreateOrganization(fmt.Sprintf("org-%s", dateTimeStr), fmt.Sprintf("coll-%s", dateTimeStr), testEmail)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Created organization %s", testOrganizationID)

	testFolderName = fmt.Sprintf("folder%s-bar", dateTimeStr)

	testCollectionID, err = webapiClient.GetCollections(testOrganizationID)
	if err != nil {
		t.Fatal(err)
	}

	areTestResourcesCreated = true

	bwClient := bwTestClient(t)
	folder, err := bwClient.CreateObject(context.Background(), bw.Object{
		Object: bw.ObjectTypeFolder,
		Name:   testFolderName,
	})
	if err != nil {
		t.Fatal(err)
	}

	testFolderID = folder.ID
	if err != nil {
		t.Fatal(err)
	}
	err = bwClient.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Synced Bitwarden client")
}

func bwTestClient(t *testing.T) bw.Client {
	vault, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	bwExec, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	client := bw.NewClient(bwExec, bw.DisableRetryBackoff(), bw.WithAppDataDir(vault))
	if len(testSessionKey) == 0 {
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
		if status.Status == bw.StatusUnauthenticated {
			err = client.LoginWithPassword(context.Background(), testEmail, testPassword)
			if err != nil {
				t.Fatal(err)
			}
		} else if status.Status == bw.StatusLocked {
			err = client.Unlock(context.Background(), testPassword)
			if err != nil {
				t.Fatal(err)
			}
		}
		testSessionKey = client.GetSessionKey()
	} else {
		client.SetSessionKey(testSessionKey)
	}
	return client
}

func tfConfigProvider() string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		session_key     = "%s"
		server          = "%s"
		email           = "%s"
		debugging_fast_mode 		= true
	}
`, testSessionKey, testServerURL, testEmail)
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
