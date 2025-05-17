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

	ensureVaultwardenHasUser(t)
	validProvider := usernamePasswordTestProvider(testEmail, testMasterPassword)
	invalidPassword := usernamePasswordTestProvider(testEmail, "incorrect-password")
	invalidAccount := usernamePasswordTestProvider("unknown-account@laverse.net", testMasterPassword)

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
	if useEmbeddedClient {
		t.Skip("Skipping test because embedded client doesn't support session key authentication")
	}
	ensureVaultwardenHasUser(t)

	validProvider := sessionKeyTestProvider(testEmail, bwOfficialTestClient(t).GetSessionKey())
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
	var useEmbeddedClientStr string
	if useEmbeddedClient {
		useEmbeddedClientStr = "true"
	} else {
		useEmbeddedClientStr = "false"
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
`, password, testServerURL, email, useEmbeddedClientStr)
}

func checkResourceId() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			"bitwarden_item_login.foo", schema_definition.AttributeID, regexp.MustCompile(regExpId),
		),
	)
}
