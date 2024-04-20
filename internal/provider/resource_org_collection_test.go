package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOrgCollection(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_org_collection.foo_org_col"
	var objectID string

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigProvider() + tfConfigResourceOrgCollection(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName, attributeName, regexp.MustCompile("^org-col-bar$"),
					),
					resource.TestMatchResourceAttr(
						resourceName, attributeID, regexp.MustCompile(regExpId),
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: orgCollectionImportID(resourceName),
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func orgCollectionImportID(resourceName string) func(s *terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		orgCollectionRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", testOrganizationID, orgCollectionRs.Primary.ID), nil
	}
}

func tfConfigResourceOrgCollection() string {
	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "org-col-bar"
}
`, testOrganizationID)
}
