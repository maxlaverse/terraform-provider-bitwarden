//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemLoginAttributes(t *testing.T) {
	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("datalogin"),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("datalogin") + tfConfigDataItemLogin(),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigInexistentDataItemLogin(),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("datalogin"),
			},
		},
	})
}

func TestAccDataSourceItemLoginFailsOnWrongResourceType(t *testing.T) {
	ensureTestConfigurationReady(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSecureNote(),
			},
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSecureNote() + tfConfigDataItemLoginCrossReference(),
				ExpectError: regexp.MustCompile("Error: returned object type does not match requested object type"),
			},
		},
	})
}

func TestAccDataSourceItemLoginBySearch(t *testing.T) {
	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("search"),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("search") + tfConfigDataItemLoginWithSearchAndOrg("test-username"),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("search") + tfConfigResourceItemLoginDuplicate() + tfConfigDataItemLoginWithSearchAndOrg("test-username"),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("search") + tfConfigResourceItemLoginDuplicate() + tfConfigDataItemLoginWithSearchOnly("test-username"),
				ExpectError: regexp.MustCompile("Error: too many objects found"),
			},
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("search") + tfConfigDataItemLoginWithSearchAndOrg("missing-item"),
				ExpectError: regexp.MustCompile("Error: no object found matching the filter"),
			},
			// Test: differentiate between items with the same username based on URL
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("search") + tfConfigResourceItemLoginDuplicate() + tfConfigDataItemLoginWithSearchAndUrl("test-username", "https://host"),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
			// Test: soft-deleting objects are not returned
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin),
			},
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigDataItemLoginWithSearchAndOrg("test-username"),
				ExpectError: regexp.MustCompile("Error: no object found matching the filter"),
			},
			// Test: search for a secure note item with a login data source should fail
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSecureNote(),
			},
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSecureNote() + tfConfigDataItemLoginWithSearchAndOrg("secure-bar"),
				ExpectError: regexp.MustCompile("Error: no object found matching the filter"),
			},
		},
	})
}

func tfConfigDataItemLoginWithSearchAndOrg(search string) string {
	return fmt.Sprintf(`
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	search = "%s"
	filter_organization_id = "%s"
}
`, search, testConfiguration.Resources.OrganizationID)
}

func tfConfigDataItemLoginWithSearchAndUrl(search, url string) string {
	return fmt.Sprintf(`
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	search = "%s"
	filter_url = "%s"
}
`, search, url)
}

func tfConfigDataItemLoginWithSearchOnly(search string) string {
	return fmt.Sprintf(`
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	search = "%s"
}
`, search)
}

func tfConfigResourceItemLoginDuplicate() string {
	return `
	resource "bitwarden_item_login" "foo_duplicate" {
		provider 			= bitwarden

		name 					= "another item with username 'test-username'"
		username 			= "test-username"
	}
	`
}

func tfConfigDataItemLogin() string {
	return `
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	id 			= bitwarden_item_login.foo.id
}
`
}

func tfConfigDataItemLoginCrossReference() string {
	return `
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	id 			= bitwarden_item_secure_note.foo.id
}
`
}

func tfConfigInexistentDataItemLogin() string {
	return `
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	id 			= 123456789
}
`
}
