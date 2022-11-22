package provider

import (
	"bytes"
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
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/executor"
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
	if err != nil && !strings.Contains(err.Error(), "User already exists") {
		t.Fatal(err)
	}
	isUserCreated = true
}

func getTestSessionKey(t *testing.T) (string, string) {
	abs, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	bwExecutable, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	env := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("BITWARDENCLI_APPDATA_DIR=%s", abs),
		"BW_NOINTERACTION=true",
		fmt.Sprintf("BW_PASSWORD=%s", testPassword),
	}

	var out bytes.Buffer

	cmd := executor.New()
	err = cmd.NewCommand(bwExecutable, "login", testEmail, "--raw", "--passwordenv", "BW_PASSWORD").WithOutput(&out).WithEnv(env).Run()
	if err != nil && !strings.Contains(err.Error(), "You are already logged in as test@laverse.net") {
		t.Fatal(err)
	}
	err = cmd.NewCommand(bwExecutable, "unlock", "--raw", "--passwordenv", "BW_PASSWORD").WithOutput(&out).WithEnv(env).Run()
	if err != nil {
		t.Fatal(err)
	}
	sessionKey := out.String()

	err = cmd.NewCommand(bwExecutable, "status", "--session", sessionKey).WithOutput(&out).WithEnv(env).Run()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), `"status":"unlocked"`) {
		t.Fatal(out.String())
	}
	return sessionKey, abs
}

func ensureVaultwardenConfigured(t *testing.T) {
	testResourcesMu.Lock()
	defer testResourcesMu.Unlock()

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
	status, err := bwClient.Status()
	if err != nil {
		t.Fatal(err)
	}

	if status.ServerURL != testServerURL || status.UserEmail != testEmail {
		if status.Status != bw.StatusUnauthenticated {
			err = bwClient.Logout()
			if err != nil {
				t.Fatal(err)
			}
		}

		err = bwClient.SetServer(testServerURL)
		if err != nil {
			t.Fatal(err)
		}

		err = bwClient.LoginWithPassword(testEmail, testPassword)
		if err != nil {
			t.Fatal(err)
		}
	}

	if !bwClient.HasSessionKey() {
		err = bwClient.Unlock(testPassword)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create a couple of test resources
	testFolderID = createTestResourceFolder(t, bwClient)
	testItemLoginID = createTestResourceLogin(t, bwClient)
	testItemSecureNoteID = createTestResourceSecureNote(t, bwClient)

	areTestResourcesCreated = true
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

func tfTestProvider() string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
	}
`, testPassword, testServerURL, testEmail)
}
