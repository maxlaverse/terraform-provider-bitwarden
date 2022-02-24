package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceItemSecureNote(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfTestProvider() + tfTestResourceItemSecureNote(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo", attributeFolderID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo", attributeID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestCheckResourceAttr(
						"bitwarden_item_secure_note.foo", attributeName, "bar",
					),
					resource.TestCheckResourceAttr(
						"bitwarden_item_secure_note.foo", attributeNotes, "notes",
					),
					resource.TestCheckResourceAttr(
						"bitwarden_item_secure_note.foo", fmt.Sprintf("%s.2.hidden", attributeField), "value-hidden",
					),
				),
			},
			{
				Config: tfTestProvider() + tfTestResourceItemSecureNoteList(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", attributeFolderID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", attributeID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestCheckResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", attributeName, "bar",
					),
					resource.TestCheckResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", attributeNotes, "notes",
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", attributeOrganizationID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", fmt.Sprintf("%s.#", attributeCollectionIDs), regexp.MustCompile("^1$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo_in_org", fmt.Sprintf("%s.0", attributeCollectionIDs), regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
				),
			},
		},
	})
}

func tfTestResourceItemSecureNote() string {
	return fmt.Sprintf(`
	resource "bitwarden_item_secure_note" "foo" {
		provider 			= bitwarden

		folder_id 			= "%s"
		name     			= "bar"
		notes 				= "notes"
		reprompt			= true

		field {
			name = "field-text"
			text = "value-text"
		}

		field {
			name    = "field-boolean"
			boolean = true
		}

		field {
			name   = "field-hidden"
			hidden = "value-hidden"
		}
	}
`, testFolderID)
}

func tfTestResourceItemSecureNoteList() string {
	return fmt.Sprintf(`
	resource "bitwarden_item_secure_note" "foo_in_org" {
		provider 			= bitwarden

		organization_id		= "%s"
		collection_ids		= ["%s"]
		folder_id 			= "%s"
		name     			= "bar"
		notes 				= "notes"
	}
`, testOrganizationID, testCollectionID, testFolderID)
}
