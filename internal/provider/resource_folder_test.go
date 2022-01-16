package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceFolder(t *testing.T) {
	ensureTestProvider(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfTestProvider() + tfTestResourceFolder(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"bitwarden_folder.foo", attributeID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_folder.foo", attributeName, regexp.MustCompile("^bar$"),
					),
				),
			},
		},
	})
}

func tfTestResourceFolder() string {
	return `
resource "bitwarden_folder" "foo" {
	provider 			= bitwarden
	name     			= "bar"
}
`
}
