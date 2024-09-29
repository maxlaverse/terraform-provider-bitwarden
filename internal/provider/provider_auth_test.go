package provider

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
)

const (
	testResource = `
	resource "bitwarden_item_login" "foo" {
		provider = bitwarden
		name     = "test"
	}
	`
)

func TestAccProviderAuthUsernamePassword(t *testing.T) {
	ensureVaultwardenHasUser(t)
	validProvider := usernamePasswordTestProvider(testEmail, testPassword)
	invalidPassword := usernamePasswordTestProvider(testEmail, "incorrect-password")
	invalidAccount := usernamePasswordTestProvider("unknown-account@laverse.net", testPassword)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidAccount + testResource,
				ExpectError: regexp.MustCompile("Username or password is incorrect"),
			}, {
				// We need to login with a valid account if we want to be able to
				// test an invalid master password, as we do bellow.
				Config: validProvider + testResource,
				Check:  checkResourceId(),
			}, {
				Config:      invalidPassword + testResource,
				ExpectError: regexp.MustCompile("Invalid master password"),
			}, {
				// We need to finish with a valid example if we don't want the TestStep to
				// fail on post-destroy.
				Config: validProvider + testResource,
				Check:  checkResourceId(),
			},
		},
	})
}

func TestAccProviderAuthSessionKey(t *testing.T) {
	ensureVaultwardenHasUser(t)

	vault, err := filepath.Abs("./.bitwarden")
	if err != nil {
		t.Fatal(err)
	}

	bwExec, err := exec.LookPath("bw")
	if err != nil {
		t.Fatal(err)
	}

	client := bwcli.NewClient(bwExec, bwcli.DisableRetryBackoff(), bwcli.WithAppDataDir(vault))
	err = client.SetServer(context.Background(), testServerURL)
	if err != nil {
		t.Fatal(err)
	}
	err = client.LoginWithPassword(context.Background(), testEmail, testPassword)
	if err != nil {
		err = client.Unlock(context.Background(), testPassword)
		if err != nil {
			t.Fatal(err)
		}
	}

	validProvider := sessionKeyTestProvider(testEmail, client.GetSessionKey())
	invalidProvider := sessionKeyTestProvider(testEmail, "invalid-session-key")

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidProvider + testResource,
				ExpectError: regexp.MustCompile("unable to unlock Vault with provided session key"),
			},
			{
				// We need to finish with a valid example if we don't want the TestStep to
				// fail on post-destroy.
				Config: validProvider + testResource,
				Check:  checkResourceId(),
			},
		},
	})
}

func sessionKeyTestProvider(email, sessionKey string) string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		server          = "%s"
		email           = "%s"
		session_key = "%s"
	}
`, testServerURL, email, sessionKey)
}

func usernamePasswordTestProvider(email, password string) string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"
	}
`, password, testServerURL, email)
}

func checkResourceId() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			"bitwarden_item_login.foo", attributeID, regexp.MustCompile(regExpId),
		),
	)
}
