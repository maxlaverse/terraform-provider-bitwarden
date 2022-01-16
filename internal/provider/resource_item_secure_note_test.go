package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceItemSecureNote(t *testing.T) {
	ensureTestProvider(t)

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
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo", attributeName, regexp.MustCompile("^bar$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_secure_note.foo", attributeNotes, regexp.MustCompile("^notes$"),
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
	}
`, testFolderID)
}
