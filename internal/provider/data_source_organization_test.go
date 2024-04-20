package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrganizationAttributes(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_organization.foo_data"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider(),
			},
			{
				Config: tfConfigProvider() + tfConfigDataOrganization(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, attributeName, regexp.MustCompile("^org-([0-9]+)-([0-9]+)-([0-9]+)$"),
					),
					resource.TestMatchResourceAttr(
						resourceName, attributeID, regexp.MustCompile(regExpId),
					),
				),
			},
		},
	})
}

func tfConfigDataOrganization() string {
	return fmt.Sprintf(`
data "bitwarden_organization" "foo_data" {
	provider	= bitwarden

	id = "%s"
}
`, testOrganizationID)
}
