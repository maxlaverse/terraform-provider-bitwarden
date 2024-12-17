package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccResourceOrgCollection(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_org_collection.foo_org_col"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeName, "org-col-bar",
					),
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-new-name-bar"),
				Check: resource.TestCheckResourceAttr(
					resourceName, schema_definition.AttributeName, "org-col-new-name-bar",
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

func tfConfigResourceOrgCollection(name string) string {
	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "%s"
}
`, testOrganizationID, name)
}
