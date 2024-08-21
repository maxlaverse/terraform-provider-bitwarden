package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFolderAttributes(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigDataFolder(),
				Check:  checkObject("data.bitwarden_folder.foo_data"),
			},
		},
	})
}

func tfConfigDataFolder() string {
	return fmt.Sprintf(`
data "bitwarden_folder" "foo_data" {
	provider	= bitwarden

	search 	= "%s"
}
`, testFolderName)
}
