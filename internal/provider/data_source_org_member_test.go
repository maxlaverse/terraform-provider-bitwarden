//go:build integration

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrgMemberAttribute(t *testing.T) {
	SkipIfOfficialCLI(t, "org members are not supported on the official CLI")

	ensureVaultwardenConfigured(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigDataOrgOwner(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "id", testAccountEmailOrgOwnerInTestOrgUserId,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "organization_id", testOrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "email", testAccountEmailOrgOwner,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_owner", "name", testAccountNameOrgOwner,
					),
				),
			}, {
				Config: tfConfigPasswordManagerProvider() + tfConfigDataOrgMembers(),
				SkipFunc: func() (bool, error) {
					return IsOfficialBackend(), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_admin", "id", testAccountEmailOrgAdminInTestOrgUserId,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_admin", "organization_id", testOrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_admin", "email", testAccountEmailOrgAdmin,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_manager", "id", testAccountEmailOrgManagerInTestOrgUserId,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_manager", "organization_id", testOrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_manager", "email", testAccountEmailOrgManager,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user", "id", testAccountEmailOrgUserInTestOrgUserId,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user", "organization_id", testOrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user", "email", testAccountEmailOrgUser,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user_by_id", "id", testAccountEmailOrgUserInTestOrgUserId,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user_by_id", "organization_id", testOrganizationID,
					),
					resource.TestCheckResourceAttr(
						"data.bitwarden_org_member.org_user_by_id", "email", testAccountEmailOrgUser,
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
		testOrganizationID, testAccountEmailOrgOwner,
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
		testOrganizationID, testAccountEmailOrgAdmin,
		testOrganizationID, testAccountEmailOrgManager,
		testOrganizationID, testAccountEmailOrgUser,
		testOrganizationID, testAccountEmailOrgUserInTestOrgUserId,
	)
}
