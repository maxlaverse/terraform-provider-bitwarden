package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceFolder(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_folder.foo"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceFolder("folder-bar"),
				Check: resource.ComposeTestCheckFunc(
					checkObject(resourceName),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceFolder("folder-new-name-bar"),
				Check: resource.ComposeTestCheckFunc(
					checkObject(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, attributeName, "folder-new-name-bar",
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     objectID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func tfConfigResourceFolder(name string) string {
	return fmt.Sprintf(`
resource "bitwarden_folder" "foo" {
	provider = bitwarden

	name     = "%s"
}
`, name)
}
