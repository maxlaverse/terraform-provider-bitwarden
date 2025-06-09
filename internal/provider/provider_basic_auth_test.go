//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
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
	SkipIfOfficialBackend(t, "Skipping test because official backend asks for a code to be sent to the email address")

	ensureTestAccountsExist(t)
	validProvider := usernamePasswordTestProvider(testConfiguration.Accounts[testAccountFullAdmin].Email, testConfiguration.Accounts[testAccountFullAdmin].Password)
	invalidPassword := usernamePasswordTestProvider(testConfiguration.Accounts[testAccountFullAdmin].Email, "incorrect-password")
	invalidAccount := usernamePasswordTestProvider("unknown-account@laverse.net", testConfiguration.Accounts[testAccountFullAdmin].Password)

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
				SkipFunc:    func() (bool, error) { return testConfiguration.UseEmbeddedClient, nil },
			}, {
				// We need to finish with a valid example if we don't want the TestStep to
				// fail on post-destroy.
				Config:   validProvider + testResource,
				Check:    checkResourceId(),
				SkipFunc: func() (bool, error) { return testConfiguration.UseEmbeddedClient, nil },
			},
		},
	})
}

func TestAccProviderAuthSessionKey(t *testing.T) {
	if testConfiguration.UseEmbeddedClient {
		t.Skip("Skipping test because embedded client doesn't support session key authentication")
	}
	ensureTestAccountsExist(t)

	validProvider := sessionKeyTestProvider(testConfiguration.Accounts[testAccountFullAdmin].Email, bwCLITestClient(t).GetSessionKey())
	invalidProvider := sessionKeyTestProvider(testConfiguration.Accounts[testAccountFullAdmin].Email, "invalid-session-key")

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
`, testConfiguration.ServerURL, email, sessionKey)
}

func usernamePasswordTestProvider(email, password string) string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		master_password = "%s"
		server          = "%s"
		email           = "%s"

		experimental {
			embedded_client = %s
		}
	}
`, password, testConfiguration.ServerURL, email, testConfiguration.UseEmbeddedClientStr())
}

func checkResourceId() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			"bitwarden_item_login.foo", schema_definition.AttributeID, regexp.MustCompile(regExpId),
		),
	)
}
