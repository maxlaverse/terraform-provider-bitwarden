//go:build integration

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrgMemberAttribute(t *testing.T) {
	ensureVaultwardenConfigured(t)

	if !useEmbeddedClient {
		t.Skip("Skipping test because official client doesn't support org members")
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigDataOrgMembers(),
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
						"data.bitwarden_org_member.org_owner", "name", fmt.Sprintf("test-%s", testUniqueIdentifier),
					),
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

func tfConfigDataOrgMembers() string {
	return fmt.Sprintf(`
data "bitwarden_org_member" "org_owner" {
	provider	= bitwarden
	organization_id = "%s"

	email = "%s"
}

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
		testOrganizationID, testAccountEmailOrgOwner,
		testOrganizationID, testAccountEmailOrgAdmin,
		testOrganizationID, testAccountEmailOrgManager,
		testOrganizationID, testAccountEmailOrgUser,
		testOrganizationID, testAccountEmailOrgUserInTestOrgUserId,
	)
}
