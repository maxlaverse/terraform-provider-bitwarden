package provider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

const (
	// Constants used to interact with a test Vaultwarden instance
	testEmail    = "test@laverse.net"
	testPassword = "test1234"

	testSignupRequest = `
	{
		"email":"test@laverse.net",
		"name":"Test",
		"masterPasswordHash":"RCLJuZZlOzRSA0ACEXSCQhF+ybeVa89Lnz7CQfeyx6Q=",
		"key":"2.A/89QpHf5lcmfJhoEHKbLw==|gUk+VUVzeaArKHGTz4O8ds0EoIiugPnTiZ59li6uy/k7lnAhantnrZtx6xVrEeNWjzaPuoWVUJ5rycESsLPYqn1eUak5OWO43tdFvD0beic=|CvPdfaaIbbsBr3Tnius/n69Rg60v/AoBQTfecWgFsv4=",
		"kdf":0,
		"kdfIterations":100000,
		"keys":{
			"publicKey":"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu/fkx3AqhZimogwqZxGde0guGsXMu4nGDhBGoxb2CTUP+WiKger7aFP/Yqx4vt2rd0913iy9o49LkXId4BhFa3hi6QBzOZoakwpBi+98ucsXARe/p0kEE2jiPT1llvVAe+yiJfnBPG1V33am8B6wHh1gfIoURb4utWSBj3QnwKK/dvBhzPqC72JHFg/BaDEYfNNA06BePCd/t/LispSZFLUT3gD2TVV0XGATl4++eJubs98hw1nFgw/Xbl+DgYVejwodGStgjAu2CrIg1PE9Ucpfuz7Oun1MJOsJ3VEebDeeJnLrj0AKYn/VjUZAw5PnOoQs3KMAj2ZAp+uj485b4wIDAQAB",
			"encryptedPrivateKey":"2.qDA5XM89wvQtaGFcpGb4xA==|Y/en3UAtRyofV9jVLi6Q+2Df8cy/bDi799fVoWIbwrD/y/Ha7bGBtret8c3tMjY7kX5BK2Xum/cKMdf1oZLk9YVLSJyviIYmnZ61m1SQ+4c5MDYHhBhpN9/eVspa4kKvYLQrzLZJh7ftRrARrTgZfRx24Op1pp/r+IEHgGwmYNcOd7OopUC/5+aTpS284mVL64WQbZJ2aDztUSuIxWh5VnvuXeM7U6Jgi0GOfPt8kKRvFrfPULT6rInZqLscZCkN390ofXgl6Hh9lsUu3xpPcmxqGO6iQEx8701HO4MGmGUyO3Ico+EKsXNKe7hQnuuujff609N6g1+BTnOgmJiBWr6mG21kaEnyNNX2zdCEmVQUrzRBa8QvslO4Z/o1dZfLNpt+Lt52OvDIpjMZAY0UmALsO8xI2dTT747Qdx+JkMqKOkjItWIhSf/gr6wU6dxI8ka12PgVY4aKtW73FbfDVwNUkHePtJVDtUWtcQkp7hYLT58tEJtjrU10cjBE8WE2wi5ImvlZeVMvYKU4Fz2a8lHQonB3pqzhXs90Qx8+ldhoL/B82Q7HSXiBiu1u0Mbia99hF1WRtnxX/66J2oaivRBLVI8FD4QhDXB3NNQFXiQnRSjpUxF24UChQxqHqqtJq1aD0LpN/8LF+elSuJ/waVrNT5a/vjHEN8G78amjtrtkVtsGooL+WFOkdh7RBaNC6MRgFo5rIZo/skfHzYrBefnU4LHDy2SV004adBajCrH271mc0Tw3q2Q9ueLCNgHdJaQk++1hYATEsw1hP1G7XNmTVsL5uDJD1yOw+8pVA/pFfdz4NEzOdvPZKq2gp5INlgv26b8Fd32VDvZmSoYr7IgJOI/F4ImjPalTwFEnxYsR+IG9wrmOTPk3xZBQiNj0dqnxSm/PkiVKhnCzSXR93Pl0d5gdrvXVwsEqAg2L5NZlRTkpij3WqkHT4fuWFfVMzUAh/V2hCQOoywu0eZwALcqRePleI8A+7K2A22f5u9blA9RmAt/ORIiiTs+1X4Wsn558eYUvjMVLrVELRvoTNMtatL18QozfDJKMXWRX/iZ0mDYU9t2HUssxXcwIywvhQFksmCRd7YuyjL4hkwOTCcQYRctjsWXR54DVl1Fr2r+12rwskQ2QUw7eZSC4wqOKb2DojDi+jqSxGnHYc65YTjNdTnJzuyYF1d1OB5qndVmaRgXqRIPdHoyYintwWXdr60oqarRIkwGnwE/KDeJgliUgdfVabk2ynSHaFOS8JTVlArf4IqKTe3De26R8+kUKkdYFHXSFM5wVm1bmJj8zA6GnMPzRJ5PqdooKbsP60uMxsgPAUYx89aNuvRAAbw4t8Az6bGbeKxQ+5BoxGkdZ2n8IVaFEzc0iC/kV9GR2wcNKbftiR8HAz7EIbAR2RTLHRI1Z+2Lf9MZrLA+KnIIRbCgEHoChdRNMrHODvOnpMxnPXjexeAVgLZLOJ9Qs363mCK2RPeJQ6DtDHQc58RZQOE9AHMNuvVNL3AJ/5wPBGXiuKQ+mlYZVIwIQp/Xs9abcYnQzTue0PI8D1vg8DVS/vTzthq6/NE2eII0o6SusDUXqBZqKDe+YiMOHZl6TsFGV5djVXqcrO+VooQL0BLqeedeHzLdH1hPExvuV97ENzOE=|ShoDi5m8hcyS78JHl4nY9ibvAloHCK2zTqypP4IXSTE="
		}
	}
	`
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
		host = "8080"
	}
	testServerURL = fmt.Sprintf("http://%s:%s/", host, port)
}

// code with undocumented assumptions and poor error handling.
// don't hesitate to ping me! (unless I fixed this first?)
func ensureTestProvider(t *testing.T) {
	mu.Lock()
	defer mu.Unlock()

	if isTestProviderConfigured {
		return
	}

	setTestServerUrl()
	createTestUser(t)

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
		Name:   "test-folder",
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
		Name:   "test-login",
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
		Name:   "test-secure-note",
		Object: bitwarden.ObjectTypeItem,
		Type:   bitwarden.ItemTypeSecureNote,
		Notes:  "Hello this is my note",
	}
	note, err := apiClient.CreateObject(newItem)
	if err != nil {
		t.Fatal(err)
	}
	return note.ID
}

func createTestUser(t *testing.T) {
	signupUrl := fmt.Sprintf("%s/api/accounts/register", testServerURL)

	resp, err := http.Post(signupUrl, "application/json", bytes.NewBufferString(testSignupRequest))
	if err != nil {
		t.Fatalf("error during registration call: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error during registration call body reading: %v", err.Error())
	}
	strBody := string(body)
	if resp.StatusCode != 200 {
		if resp.StatusCode == 400 && strings.Contains(strBody, "User already exists") {
			return
		}
		t.Fatalf("bad status code for registration call: %d, %s", resp.StatusCode, strBody)
	}
}
