package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemLogin(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_item_login.foo_data"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemLogin(),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemLogin() + tfConfigDataItemLogin(),
				Check:  checkItemLogin(resourceName),
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
