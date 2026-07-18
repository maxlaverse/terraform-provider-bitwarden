//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccResourceOrgGroup(t *testing.T) {
	SkipIfOfficialBackend(t, "org groups require a higher license to be tested")
	SkipIfOfficialCLI(t, "org groups are not supported by the official CLI")

	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_org_group.foo_group"
	groupName := fmt.Sprintf("acc-group-%s", testConfiguration.UniqueTestIdentifier)
	renamedGroupName := groupName + "-renamed"
	memberID := testConfiguration.Accounts[testAccountOrgUser].UserIdInTestOrganization
	member2ID := testConfiguration.Accounts[testAccountOrgManager].UserIdInTestOrganization

	var originalID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// Create with no members
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgGroup(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeName, groupName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "0"),
					resource.TestMatchResourceAttr(resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId)),
					getObjectID(resourceName, &originalID),
				),
			},
			// Add one member in-place
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgGroup(groupName, memberID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeName, groupName),
					resource.TestCheckResourceAttr(resourceName, "member.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "member.*", memberID),
					checkSameID(resourceName, &originalID),
				),
			},
			// Add a second member in-place
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgGroup(groupName, memberID, member2ID) + tfConfigDataOrgGroupReadFromResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "member.*", memberID),
					resource.TestCheckTypeSetElemAttr(resourceName, "member.*", member2ID),
					resource.TestCheckResourceAttr("data.bitwarden_org_group.foo_group_read", "member.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.bitwarden_org_group.foo_group_read", "member.*", memberID),
					resource.TestCheckTypeSetElemAttr("data.bitwarden_org_group.foo_group_read", "member.*", member2ID),
					checkSameID(resourceName, &originalID),
				),
			},
			// Remove all members in-place
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgGroup(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "member.#", "0"),
					checkSameID(resourceName, &originalID),
				),
			},
			// Renaming forces replacement
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceOrgGroup(renamedGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeName, renamedGroupName),
					checkChangedID(resourceName, &originalID),
				),
			},
			// Import
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: orgGroupImportID(resourceName),
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func tfConfigResourceOrgGroup(groupName string, memberIDs ...string) string {
	memberLines := ""
	if len(memberIDs) > 0 {
		quoted := make([]string, len(memberIDs))
		for i, id := range memberIDs {
			quoted[i] = fmt.Sprintf(`"%s"`, id)
		}
		memberLines = fmt.Sprintf("\n\tmember = [%s]\n", strings.Join(quoted, ", "))
	}
	return fmt.Sprintf(`
resource "bitwarden_org_group" "foo_group" {
	organization_id = "%s"
	name            = "%s"%s
}
`, testConfiguration.Resources.OrganizationID, groupName, memberLines)
}

func orgGroupImportID(resourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", testConfiguration.Resources.OrganizationID, rs.Primary.ID), nil
	}
}

func checkSameID(resourceName string, original *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID != *original {
			return fmt.Errorf("expected same ID %s, got %s", *original, rs.Primary.ID)
		}
		return nil
	}
}

func checkChangedID(resourceName string, original *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == *original {
			return fmt.Errorf("expected ID to have changed, got %s", rs.Primary.ID)
		}
		return nil
	}
}

func tfConfigDataOrgGroupReadFromResource() string {
	return fmt.Sprintf(`
data "bitwarden_org_group" "foo_group_read" {
	organization_id = "%s"
	filter_name     = bitwarden_org_group.foo_group.name
}
`, testConfiguration.Resources.OrganizationID)
}
