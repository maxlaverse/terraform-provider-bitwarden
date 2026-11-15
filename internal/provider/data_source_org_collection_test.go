//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccDataSourceOrgCollection(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "data.bitwarden_org_collection.foo_data"

	var nameAssertion resource.TestCheckFunc
	if IsOfficialBackend() {
		nameAssertion = resource.TestCheckResourceAttr(
			resourceName, schema_definition.AttributeName, "Default collection",
		)
	} else {
		nameAssertion = resource.TestCheckResourceAttr(
			resourceName, schema_definition.AttributeName, "coll-"+testConfiguration.UniqueTestIdentifier,
		)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigDataOrgCollection(),
				Check: resource.ComposeTestCheckFunc(
					nameAssertion,
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
					),
				),
			},
		},
	})
}

func tfConfigDataOrgCollection() string {
	search := "Default collection"
	if IsVaultwardenBackend() {
		search = "coll-" + testConfiguration.UniqueTestIdentifier
	}
	return fmt.Sprintf(`
data "bitwarden_org_collection" "foo_data" {
	provider	= bitwarden

	organization_id = "%s"

	search 	= "%s"
}
`, testConfiguration.Resources.OrganizationID, search)
}
