//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccDataSourceOrgCollection(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_org_collection.foo_data"

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigDataOrgCollection(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeName, regexp.MustCompile("^coll-([0-9]{6})$"),
					),
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
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
