package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	resource.UnitTest(t, resource.TestCase{
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
	sessionKey, vaultPath := getTestSessionKey(t)
	validProvider := sessionKeyTestProvider(testEmail, vaultPath, sessionKey)
	invalidSessionKey := sessionKeyTestProvider(testEmail, vaultPath, "invalid-session-key")

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      invalidSessionKey + testResource,
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

func sessionKeyTestProvider(email, vaultPath, sessionKey string) string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		server          = "%s"
		email           = "%s"
		vault_path = "%s"
		session_key = "%s"
	}
`, testServerURL, email, vaultPath, sessionKey)
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
			"bitwarden_item_login.foo", attributeID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
		),
	)
}
