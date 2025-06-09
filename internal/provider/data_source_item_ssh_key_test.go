//go:build integration

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemSSHKey(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "data.bitwarden_item_ssh_key.foo_data"

	resource.Test(t, resource.TestCase{
		ProviderFactories:         providerFactories,
		PreventPostDestroyRefresh: false,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSSHKey(),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSSHKey() + tfConfigDataItemSSHKey(),
				Check:  checkItemSSHKey(resourceName),
			},
		},
	})
}

func tfConfigDataItemSSHKey() string {
	return `
data "bitwarden_item_ssh_key" "foo_data" {
	provider 	= bitwarden

	id 			= bitwarden_item_ssh_key.foo.id
}
`
}
