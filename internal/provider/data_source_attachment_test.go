//go:build integration

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccDataSourceAttachment(t *testing.T) {
	SkipIfNonPremiumTestAccount(t)

	ensureVaultwardenConfigured(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigAttachmentSpecificPasswordManagerProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
			},
			{
				Config: tfConfigAttachmentSpecificPasswordManagerProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachment(),
				Check: resource.TestMatchResourceAttr(
					"data.bitwarden_attachment.foo_data", schema_definition.AttributeAttachmentContent, regexp.MustCompile(`^Hello, I'm a text attachment$`),
				),
			},
			{
				Config:      tfConfigAttachmentSpecificPasswordManagerProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachmentInexistent(),
				ExpectError: regexp.MustCompile("Error: attachment not found"),
			},
			{
				Config:      tfConfigAttachmentSpecificPasswordManagerProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt") + tfConfigDataAttachmentInexistentItem(),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
			{
				Config: tfConfigAttachmentSpecificPasswordManagerProvider() + tfConfigResourceOrganizationAttachment("fixtures/attachment1.txt", testOrganizationID),
				SkipFunc: func() (bool, error) {
					// Organization attachments are not support with 'free' and 'premium' plans.
					return IsOfficialBackend(), nil
				},
			},
			{
				Config: tfConfigAttachmentSpecificPasswordManagerProvider() + tfConfigResourceOrganizationAttachment("fixtures/attachment1.txt", testOrganizationID) + tfConfigDataAttachment(),
				Check: resource.TestMatchResourceAttr(
					"data.bitwarden_attachment.foo_data", schema_definition.AttributeAttachmentContent, regexp.MustCompile(`^Hello, I'm a text attachment$`),
				),
				SkipFunc: func() (bool, error) {
					// Organization attachments are not support with 'free' and 'premium' plans.
					return IsOfficialBackend(), nil
				},
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
	item_id 	= "71767b68-b385-4440-878c-b2e500bfcff9"
}
`
}
