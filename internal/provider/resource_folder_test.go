package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceFolder(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_folder.foo"
	var objectID string

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceFolder(),
				Check: resource.ComposeTestCheckFunc(
					checkObject(resourceName),
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

func tfConfigResourceFolder() string {
	return `
resource "bitwarden_folder" "foo" {
	provider = bitwarden

	name     = "folder-bar"
}
`
}
