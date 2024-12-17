package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
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
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceFolder("folder-bar"),
				Check: resource.ComposeTestCheckFunc(
					checkObject(resourceName),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceFolder("folder-new-name-bar"),
				Check: resource.ComposeTestCheckFunc(
					checkObject(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeName, "folder-new-name-bar",
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
