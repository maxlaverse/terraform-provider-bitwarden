//go:build integration

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFolderAttributes(t *testing.T) {
	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceFolder("folder-bar"),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceFolder("folder-bar") + tfConfigDataFolder(),
				Check:  checkObject("data.bitwarden_folder.foo_data"),
			},
		},
	})
}

func tfConfigDataFolder() string {
	return `
data "bitwarden_folder" "foo_data" {
	provider	= bitwarden

	search 	= "folder-bar"
}
`
}
