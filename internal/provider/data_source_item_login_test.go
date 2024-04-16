package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
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

func TestAccDataSourceItemLoginDeleted(t *testing.T) {
	var objectID string

	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceItemLoginSmall(),
				Check:  getObjectID("bitwarden_item_login.foo", &objectID),
			},
			{
				Config: tfConfigProvider() + tfConfigDataItemLoginWithURLFilter("https://start_with/something"),
				Check:  getObjectID("bitwarden_item_login.foo", &objectID),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceItemLoginSmall() + tfConfigDataItemLoginWithId(objectID),
				PreConfig: func() {
					err := bwTestClient(t).DeleteObject("item", objectID)
					assert.NoError(t, err)
				},
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
		},
	})
}

func tfConfigDataItemLoginWithId(id string) string {
	return fmt.Sprintf(`
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	id 			= "%s"
}
`, id)
}

func tfConfigDataItemLoginWithURLFilter(url string) string {
	return fmt.Sprintf(`
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	url = "%s"
}
`, url)
}

func tfConfigDataItemLogin() string {
	return `
data "bitwarden_item_login" "foo_data" {
	provider	= bitwarden

	id 			= bitwarden_item_login.foo.id
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
