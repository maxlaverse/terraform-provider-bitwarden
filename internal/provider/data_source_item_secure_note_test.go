package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemSecureNote(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_item_secure_note.foo_data"

	resource.Test(t, resource.TestCase{
		ProviderFactories:         providerFactories,
		PreventPostDestroyRefresh: false,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemSecureNote(),
			},
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemSecureNote() + tfConfigDataItemSecureNote(),
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
