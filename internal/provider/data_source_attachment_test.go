package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAttachment(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_attachment.foo_data"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfConfigProvider() + tfConfigResourceAttachment("non-existent") + tfConfigDataAttachment(),
				ExpectError: regexp.MustCompile("no such file or directory"),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachment(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, attributeAttachmentContent, regexp.MustCompile(`^Hello, I'm a text attachment$`),
					),
				),
			},
		},
	})
}

func tfConfigDataAttachment() string {
	return `
data "bitwarden_attachment" "foo_data" {
	provider	= bitwarden

	id 			= bitwarden_attachment.foo.id
	item_id 	= bitwarden_attachment.foo.item_id
}
`
}
