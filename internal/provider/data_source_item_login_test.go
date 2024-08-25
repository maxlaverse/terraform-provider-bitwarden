package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemLoginAttributes(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceItemLogin(),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceItemLogin() + tfConfigDataItemLogin(),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
		},
	})
}

func TestAccDataSourceItemLoginFailsOnInexistentItem(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfConfigProvider() + tfConfigInexistentDataItemLogin(),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
		},
	})
}

func TestAccDataSourceItemLoginFailsOnWrongResourceType(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceItemSecureNote(),
			},
			{
				Config:      tfConfigProvider() + tfConfigResourceItemSecureNote() + tfConfigDataItemLoginCrossReference(),
				ExpectError: regexp.MustCompile("Error: returned object type does not match requested object type"),
			},
		},
	})
}

func TestAccDataSourceItemLoginBySearch(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceItemLogin() + tfConfigDataItemLoginWithSearchAndOrg("test-username"),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceItemLogin() + tfConfigResourceItemLoginDuplicate() + tfConfigDataItemLoginWithSearchAndOrg("test-username"),
				Check:  checkItemLogin("data.bitwarden_item_login.foo_data"),
			},
			{
				Config:      tfConfigProvider() + tfConfigResourceItemLogin() + tfConfigResourceItemLoginDuplicate() + tfConfigDataItemLoginWithSearchOnly("test-username"),
				ExpectError: regexp.MustCompile("Error: too many objects found"),
			},
			{
				Config:      tfConfigProvider() + tfConfigResourceItemLogin() + tfConfigDataItemLoginWithSearchAndOrg("missing-item"),
				ExpectError: regexp.MustCompile("Error: no object found matching the filter"),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceItemSecureNote(),
			},
			{
				Config:      tfConfigProvider() + tfConfigResourceItemSecureNote() + tfConfigDataItemLoginWithSearchAndOrg("secure-bar"),
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
`, search, testOrganizationID)
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
