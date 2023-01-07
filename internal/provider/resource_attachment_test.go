package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceAttachment(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_attachment.foo"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("non-existent"),
				ExpectError:  regexp.MustCompile("no such file or directory"),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, attributeAttachmentFile, regexp.MustCompile(`^34945801b5aed4540ccfde8320ec7c395325e02d$`),
					),
					resource.TestMatchResourceAttr(
						resourceName, attributeAttachmentItemID, regexp.MustCompile(regExpId),
					),
					checkAttachmentMatches(resourceName, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: attachmentImportID(resourceName, "bitwarden_item_login.foo"),
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccResourceItemAttachmentFields(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_item_login.foo"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, "attachments.#", regexp.MustCompile("^1$"),
					),
					checkAttachmentMatches(resourceName, "attachments.0."),
				),
			},
		},
	})
}

func checkAttachmentMatches(resourceName, baseAttribute string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s%s", baseAttribute, attributeID), regexp.MustCompile("^[a-fA-F0-9]{20}$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s%s", baseAttribute, attributeAttachmentFileName), regexp.MustCompile(`^attachment1.txt$`),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s%s", baseAttribute, attributeAttachmentSize), regexp.MustCompile("^81$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s%s", baseAttribute, attributeAttachmentSizeName), regexp.MustCompile(`^81\.00 bytes$`),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s%s", baseAttribute, attributeAttachmentURL), regexp.MustCompile(`^http://127.0.0.1:8080/attachments/[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}/[a-fA-F0-9]{20}$`),
		),
	)
}

func TestAccResourceItemAttachmentFileChanges(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_attachment.foo"

	var ID string
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment1.txt"),
				Check: resource.ComposeTestCheckFunc(
					compareIdentifier(resourceName, &ID, true),
					resource.TestMatchResourceAttr(
						resourceName, attributeAttachmentFile, regexp.MustCompile(`^34945801b5aed4540ccfde8320ec7c395325e02d$`),
					),
				),
			},
			{
				// Same content, different filename
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment2a.txt"),
				Check: resource.ComposeTestCheckFunc(
					compareIdentifier(resourceName, &ID, false),
					resource.TestMatchResourceAttr(
						resourceName, attributeAttachmentFile, regexp.MustCompile(`^34945801b5aed4540ccfde8320ec7c395325e02d$`),
					),
				),
			},
			{
				// Different content
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceAttachment("fixtures/attachment2b.txt"),
				Check: resource.ComposeTestCheckFunc(
					compareIdentifier(resourceName, &ID, true),
					resource.TestMatchResourceAttr(
						resourceName, attributeAttachmentFile, regexp.MustCompile(`^5d80a5115d21ca330f0d60e355ed829526dcbb47$`),
					),
				),
			},
		},
	})
}

func attachmentImportID(resourceName, resourceItemName string) func(s *terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		attachmentRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		itemRs, ok := s.RootModule().Resources[resourceItemName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceItemName)
		}

		return fmt.Sprintf("%s/%s", attachmentRs.Primary.ID, itemRs.Primary.ID), nil
	}
}

func compareIdentifier(resourceName string, ID *string, expectChange bool) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		attachmentRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		previousId := *ID
		*ID = attachmentRs.Primary.ID
		if expectChange && previousId == attachmentRs.Primary.ID {
			return fmt.Errorf("identifier didn't change! %s", attachmentRs.Primary.ID)
		}
		if !expectChange && previousId != attachmentRs.Primary.ID {
			return fmt.Errorf("identifier changed! Before: %s, After: %s", attachmentRs.Primary.ID, *ID)
		}
		return nil
	}
}

func tfConfigResourceAttachment(filepath string) string {
	return `
resource "bitwarden_item_login" "foo" {
	provider = bitwarden

	name     = "foo"
}

resource "bitwarden_attachment" "foo" {
	provider  = bitwarden

	file      = "` + filepath + `"
	item_id   = bitwarden_item_login.foo.id
}
`
}
