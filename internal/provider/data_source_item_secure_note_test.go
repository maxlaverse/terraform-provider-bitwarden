//go:build integration

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceItemSecureNote(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "data.bitwarden_item_secure_note.foo_data"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories:  providerFactories,
		PreventPostDestroyRefresh: false,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSecureNote(),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSecureNote() + tfConfigDataItemSecureNote(),
				Check:  checkItemSecureNote(resourceName),
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
