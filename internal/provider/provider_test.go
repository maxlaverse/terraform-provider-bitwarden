package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	// Constants used to interact with a test Vaultwarden instance
	testEmail     = "test@laverse.net"
	testPassword  = "test1234"
	kdfIterations = 100000
)

// Generated resources used for testing
var testServerURL string
var testFolderID string
var testItemLoginID string
var testItemSecureNoteID string
var testOrganizationID string
var testCollectionID string

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"bitwarden": func() (*schema.Provider, error) {
		return New("dev")(), nil
	},
}

var isTestProviderConfigured bool
var mu sync.Mutex

func tfTestProvider() string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
	}
`, testPassword, testServerURL, testEmail)
}

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func setTestServerUrl() {
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

func ensureVaultwardenConfigured(t *testing.T) {
	mu.Lock()
	defer mu.Unlock()

	if isTestProviderConfigured {
		return
	}

	setTestServerUrl()

	webapiClient := webapi.NewClient(testServerURL)

	userAlreadyExists := false
	err := webapiClient.RegisterUser("test", testEmail, testPassword, kdfIterations)
	if err != nil && !strings.Contains(err.Error(), "User already exists") {
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
	testOrganizationID, err = webapiClient.CreateOrganization(fmt.Sprintf("org-%d", time.Now().Unix()), fmt.Sprintf("coll-%d", time.Now().Unix()), testEmail)
	if err != nil {
		t.Fatal(err)
	}

	testCollectionID, err = webapiClient.GetCollections(testOrganizationID)
	if err != nil {
		t.Fatal(err)
	}

	abs, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	bwExecutable, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	bwClient := bw.NewClient(bwExecutable, bw.WithAppDataDir(abs))
	bwClient.SetServer(testServerURL)
	bwClient.LoginWithPassword(testEmail, testPassword)
	if !bwClient.HasSessionKey() {
		bwClient.Unlock(testPassword)
	}

	// Create a couple of test resources
	testFolderID = createTestResourceFolder(t, bwClient)
	testItemLoginID = createTestResourceLogin(t, bwClient)
	testItemSecureNoteID = createTestResourceSecureNote(t, bwClient)

	isTestProviderConfigured = true
}

func createTestResourceFolder(t *testing.T, bwClient bw.Client) string {
	newItem := bw.Object{
		Name:   fmt.Sprintf("folder-%d", time.Now().Unix()),
		Object: bw.ObjectTypeFolder,
	}
	folder, err := bwClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return folder.ID
}

func createTestResourceLogin(t *testing.T, bwClient bw.Client) string {
	newItem := bw.Object{
		Name:   fmt.Sprintf("login-%d", time.Now().Unix()),
		Object: bw.ObjectTypeItem,
		Type:   bw.ItemTypeLogin,
		Login: bw.Login{
			Username: "test-user",
			Password: "test-password",
		},
	}
	login, err := bwClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return login.ID
}

func createTestResourceSecureNote(t *testing.T, bwClient bw.Client) string {
	newItem := bw.Object{
		Name:   fmt.Sprintf("secure-note-%d", time.Now().Unix()),
		Object: bw.ObjectTypeItem,
		Type:   bw.ItemTypeSecureNote,
		Notes:  "Hello this is my note",
		Fields: []bw.Field{
			{
				Name:  "field-1",
				Value: "value-1",
				Type:  bw.FieldTypeText,
			},
			{
				Name:  "field-2",
				Value: "value-2",
				Type:  bw.FieldTypeHidden,
			},
		},
	}
	note, err := bwClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return note.ID
}
