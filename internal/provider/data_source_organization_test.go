//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccDataSourceOrganizationAttributes(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "data.bitwarden_organization.foo_data"

	var nameAssertion resource.TestCheckFunc
	if IsOfficialBackend() {
		nameAssertion = resource.TestCheckResourceAttr(
			resourceName, schema_definition.AttributeName, "MyTestOrganization",
		)
	} else {
		nameAssertion = resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeName, regexp.MustCompile("^org-([0-9]{6})$"),
		)
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigDataOrganization(),
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

func tfConfigDataOrganization() string {
	return fmt.Sprintf(`
data "bitwarden_organization" "foo_data" {
	provider	= bitwarden

	id = "%s"
}
`, testOrganizationID)
}
