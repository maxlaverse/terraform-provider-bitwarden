package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrgCollectionAttributes(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_org_collection.foo_data"

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigProvider() + tfConfigDataOrgCollection(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, attributeName, regexp.MustCompile("^coll-([0-9]{6})$"),
					),
					resource.TestMatchResourceAttr(
						resourceName, attributeID, regexp.MustCompile(regExpId),
					),
				),
			},
		},
	})
}

func tfConfigDataOrgCollection() string {
	return fmt.Sprintf(`
data "bitwarden_org_collection" "foo_data" {
	provider	= bitwarden

	organization_id = "%s"

	search 	= "col"
}
`, testOrganizationID)
}
