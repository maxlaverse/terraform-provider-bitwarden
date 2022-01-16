package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemSecureNote(t *testing.T) {
	ensureTestProvider(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfTestProvider() + tfTestDataItemSecureNote(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.bitwarden_item_secure_note.foo", attributeName, regexp.MustCompile("^test-secure-note$"),
					),
					resource.TestMatchResourceAttr(
						"data.bitwarden_item_secure_note.foo", attributeNotes, regexp.MustCompile("^Hello this is my note$"),
					),
				),
			},
		},
	})
}

func tfTestDataItemSecureNote() string {
	return fmt.Sprintf(`
data "bitwarden_item_secure_note" "foo" {
  id = "%s"
}
`, testItemSecureNoteID)
}
