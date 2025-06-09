//go:build integration

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrgMemberAttribute(t *testing.T) {
	SkipIfOfficialCLI(t, "org members are not supported on the official CLI")

	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigDataOrgOwner(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "id", testConfiguration.Accounts[testAccountOrgOwner].UserIdInTestOrganization,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "organization_id", testConfiguration.Resources.OrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "email", testConfiguration.Accounts[testAccountOrgOwner].Email,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "name", testConfiguration.Accounts[testAccountOrgOwner].Name,
					),
				),
			}, {
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigDataOrgMembers(),
				SkipFunc: func() (bool, error) {
					return IsOfficialBackend(), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_admin", "id", testConfiguration.Accounts[testAccountOrgAdmin].UserIdInTestOrganization,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_admin", "organization_id", testConfiguration.Resources.OrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_admin", "email", testConfiguration.Accounts[testAccountOrgAdmin].Email,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_manager", "id", testConfiguration.Accounts[testAccountOrgManager].UserIdInTestOrganization,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_manager", "organization_id", testConfiguration.Resources.OrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_manager", "email", testConfiguration.Accounts[testAccountOrgManager].Email,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user", "id", testConfiguration.Accounts[testAccountOrgUser].UserIdInTestOrganization,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user", "organization_id", testConfiguration.Resources.OrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user", "email", testConfiguration.Accounts[testAccountOrgUser].Email,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user_by_id", "id", testConfiguration.Accounts[testAccountOrgUser].UserIdInTestOrganization,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user_by_id", "organization_id", testConfiguration.Resources.OrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user_by_id", "email", testConfiguration.Accounts[testAccountOrgUser].Email,
					),
				),
			},
		},
	})
}

func tfConfigDataOrgOwner() string {
	return fmt.Sprintf(`
data "bitwarden_org_member" "org_owner" {
	provider	= bitwarden
	organization_id = "%s"

	email = "%s"
}

`,
		testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[testAccountOrgOwner].Email,
	)
}

func tfConfigDataOrgMembers() string {
	return fmt.Sprintf(`
data "bitwarden_org_member" "org_admin" {
	provider	= bitwarden
	organization_id = "%s"

	email = "%s"
}

data "bitwarden_org_member" "org_manager" {
	provider	= bitwarden
	organization_id = "%s"

	email = "%s"
}

data "bitwarden_org_member" "org_user" {
	provider	= bitwarden
	organization_id = "%s"

	email = "%s"
}

data "bitwarden_org_member" "org_user_by_id" {
	provider	= bitwarden
	organization_id = "%s"

	id = "%s"
}

`,
		testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[testAccountOrgAdmin].Email,
		testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[testAccountOrgManager].Email,
		testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[testAccountOrgUser].Email,
		testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[testAccountOrgUser].UserIdInTestOrganization,
	)
}
