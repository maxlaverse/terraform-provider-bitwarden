//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccResourceOrgMember(t *testing.T) {
	SkipIfOfficialBackend(t, "would accumulate orphaned invitations in the shared official test organization")
	SkipIfOfficialCLI(t, "org members are not supported by the official CLI")

	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_org_member.foo_member"
	testEmail := fmt.Sprintf("member-acc-%s@laverse.net", testConfiguration.UniqueTestIdentifier)

	var originalID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgMember(testEmail, "manager"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeEmail, testEmail),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeRole, "manager"),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeOrganizationID, testConfiguration.Resources.OrganizationID),
					resource.TestMatchResourceAttr(resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId)),
					getObjectID(resourceName, &originalID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgMember(testEmail, "admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeRole, "admin"),
					checkChangedID(resourceName, &originalID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: orgMemberImportID(resourceName),
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func tfConfigResourceOrgMember(email, role string) string {
	return fmt.Sprintf(`
resource "bitwarden_org_member" "foo_member" {
	organization_id = "%s"
	email           = "%s"
	role            = "%s"
}
`, testConfiguration.Resources.OrganizationID, email, role)
}

func orgMemberImportID(resourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", testConfiguration.Resources.OrganizationID, rs.Primary.ID), nil
	}
}
