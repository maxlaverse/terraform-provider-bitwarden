package provider

import (
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
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemLogin(),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemLogin() + tfConfigDataItemLogin(),
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
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemSecureNote(),
			},
			{
				Config:      tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemSecureNote() + tfConfigDataItemLoginCrossReference(),
				ExpectError: regexp.MustCompile("Error: returned object type does not match requested object type"),
			},
		},
	})
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
