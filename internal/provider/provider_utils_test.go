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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
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

	_, err = command.New(bwExecutable, "login", testEmail, "--raw", "--passwordenv", "BW_PASSWORD").WithOutput(&out).AppendEnv(env).Run()
	if err != nil && !strings.Contains(err.Error(), "You are already logged in as test@laverse.net") {
		t.Fatal(err)
	}
	_, err = command.New(bwExecutable, "unlock", "--raw", "--passwordenv", "BW_PASSWORD").WithOutput(&out).AppendEnv(env).Run()
	if err != nil {
		t.Fatal(err)
	}
	sessionKey := out.String()

	_, err = command.New(bwExecutable, "status", "--session", sessionKey).WithOutput(&out).AppendEnv(env).Run()
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

	areTestResourcesCreated = true
}

func tfConfigProvider() string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
	}
`, testPassword, testServerURL, testEmail)
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
