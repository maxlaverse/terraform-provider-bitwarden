package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemSecureNote(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_item_secure_note.foo_data"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemSecureNote(),
			},
			{
				Config: tfConfigProvider() + tfConfigResourceFolder() + tfConfigResourceItemSecureNote() + tfConfigDataItemSecureNote(),
				Check:  checkItemGeneral(resourceName),
			},
		},
	})
}

func tfConfigDataItemSecureNote() string {
	return `
data "bitwarden_item_secure_note" "foo_data" {
	provider 	= bitwarden

	id 			= bitwarden_item_secure_note.foo.id
}
`
}
