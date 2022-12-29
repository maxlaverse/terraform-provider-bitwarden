package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceItemAttachment(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_item_attachment.foo"
	var objectID string

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceItemAttachment(),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     objectID,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func tfConfigResourceItemAttachment() string {
	return `
resource "bitwarden_item_login" "foo" {
	provider = bitwarden

	name     = "foo"
}
	
resource "bitwarden_item_attachment" "foo" {
	provider  = bitwarden

	file      = "./resource_item_attachment.go"
	item_id   = bitwarden_item_login.foo.id
}
`
}
