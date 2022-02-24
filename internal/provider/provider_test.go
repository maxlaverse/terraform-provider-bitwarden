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
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/provider/test"
)

const (
	// Constants used to interact with a test Vaultwarden instance
	testEmail    = "test@laverse.net"
	testPassword = "test1234"
)

// Generated resources used for testing
var testServerURL string
var testFolderID string
var testItemLoginID string
var testItemSecureNoteID string

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

// code with undocumented assumptions and poor error handling.
// don't hesitate to ping me! (unless I fixed this first?)
func ensureVaultwardenConfigured(t *testing.T) {
	mu.Lock()
	defer mu.Unlock()

	if isTestProviderConfigured {
		return
	}

	setTestServerUrl()

	testClient := test.NewVaultwardenTestClient(testServerURL)
	err := testClient.RegisterUser("test", testEmail, testPassword, 100000)
	if err != nil && !strings.Contains(err.Error(), "User already exists") {
		t.Fatal(err)
	}

	abs, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	// Configure client
	opts := []bitwarden.Options{}
	opts = append(opts, bitwarden.WithAppDataDir(abs))

	bwExecutable, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	apiClient := bitwarden.NewClient(bwExecutable, opts...)
	apiClient.SetServer(testServerURL)
	apiClient.LoginWithPassword(testEmail, testPassword)
	if !apiClient.HasSessionKey() {
		apiClient.Unlock(testPassword)
	}

	// Create a couple of test resources
	testFolderID = createTestResourceFolder(t, apiClient)
	testItemLoginID = createTestResourceLogin(t, apiClient)
	testItemSecureNoteID = createTestResourceSecureNote(t, apiClient)

	isTestProviderConfigured = true
}

func createTestResourceFolder(t *testing.T, apiClient bitwarden.Client) string {
	newItem := bitwarden.Object{
		Name:   fmt.Sprintf("folder-%d", time.Now().Unix()),
		Object: bitwarden.ObjectTypeFolder,
	}
	folder, err := apiClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return folder.ID
}

func createTestResourceLogin(t *testing.T, apiClient bitwarden.Client) string {
	newItem := bitwarden.Object{
		Name:   fmt.Sprintf("login-%d", time.Now().Unix()),
		Object: bitwarden.ObjectTypeItem,
		Type:   bitwarden.ItemTypeLogin,
		Login: bitwarden.Login{
			Username: "test-user",
			Password: "test-password",
		},
	}
	login, err := apiClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return login.ID
}

func createTestResourceSecureNote(t *testing.T, apiClient bitwarden.Client) string {
	newItem := bitwarden.Object{
		Name:   fmt.Sprintf("secure-note-%d", time.Now().Unix()),
		Object: bitwarden.ObjectTypeItem,
		Type:   bitwarden.ItemTypeSecureNote,
		Notes:  "Hello this is my note",
		Fields: []bitwarden.Field{
			{
				Name:  "field-1",
				Value: "value-1",
				Type:  bitwarden.FieldTypeText,
			},
			{
				Name:  "field-2",
				Value: "value-2",
				Type:  bitwarden.FieldTypeHidden,
			},
		},
	}
	note, err := apiClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return note.ID
}
