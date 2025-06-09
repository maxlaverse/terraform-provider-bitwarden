//go:build integration

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrgGroupAttribute(t *testing.T) {
	SkipIfOfficialBackend(t, "org groups require a higher license to be tested")
	SkipIfOfficialCLI(t, "org groups are not supported by the official CLI")

	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigDataOrgGroup(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_group.default_group", "id", testConfiguration.Resources.GroupID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_group.default_group", "organization_id", testConfiguration.Resources.OrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_group.default_group", "name", testConfiguration.Resources.GroupName,
					),
				),
			},
		},
	})
}

func tfConfigDataOrgGroup() string {
	return fmt.Sprintf(`
data "bitwarden_org_group" "default_group" {
	provider	= bitwarden
	organization_id = "%s"

	filter_name = "%s"
}

`,
		testConfiguration.Resources.OrganizationID, testConfiguration.Resources.GroupName,
	)
}
