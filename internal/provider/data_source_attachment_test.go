package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAttachmentAttributes(t *testing.T) {
	t.Skip("Temporarily skipping")
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachment(),
				Check: resource.TestMatchResourceAttr(
					"data.bitwarden_attachment.foo_data", attributeAttachmentContent, regexp.MustCompile(`^Hello, I'm a text attachment$`),
				),
			},
		},
	})
}

func TestAccDataSourceAttachmentFailsOnInvalidConfiguration(t *testing.T) {
	t.Skip("Temporarily skipping")
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachmentInexistent(),
				ExpectError: regexp.MustCompile("Error: attachment not found"),
			},
			{
				Config:      tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachmentInexistentItem(),
				ExpectError: regexp.MustCompile("Error: object not found"),
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

func tfConfigDataAttachmentInexistent() string {
	return `
data "bitwarden_attachment" "foo_data" {
	provider	= bitwarden

	id 			= 0123456789
	item_id 	= bitwarden_attachment.foo.item_id
}
`
}

func tfConfigDataAttachmentInexistentItem() string {
	return `
data "bitwarden_attachment" "foo_data" {
	provider	= bitwarden

	id 			= bitwarden_attachment.foo.id
	item_id 	= 0123456789
}
`
}
