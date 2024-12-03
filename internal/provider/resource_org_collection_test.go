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

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, attributeName, "org-col-bar",
					),
					resource.TestMatchResourceAttr(
						resourceName, attributeID, regexp.MustCompile(regExpId),
					),
					resource.TestCheckResourceAttr(
						resourceName, "member.#", "2",
					),
					resource.TestMatchResourceAttr(
						resourceName, "member.0.org_user_id", regexp.MustCompile(regExpId),
					),
					resource.TestCheckResourceAttr(
						resourceName, "member.0.read_only", "false",
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, attributeName, "org-col-bar",
					),
					resource.TestMatchResourceAttr(
						resourceName, attributeID, regexp.MustCompile(regExpId),
					),
					resource.TestCheckResourceAttr(
						resourceName, "member.#", "2",
					),
					resource.TestMatchResourceAttr(
						resourceName, "member.0.org_user_id", regexp.MustCompile(regExpId),
					),
					resource.TestCheckResourceAttr(
						resourceName, "member.0.read_only", "false",
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-new-name-bar"),
				Check: resource.TestCheckResourceAttr(
					resourceName, attributeName, "org-col-new-name-bar",
				),
			},
			// {
			// 	ResourceName: resourceName,
			// 	Config:       tfConfigPasswordManagerProvider() + tfProviderOrgOwner(testAccountEmailOrgOwner) + tfConfigResourceOrgCollection("org-col-new-name-bar"),
			// 	Check: resource.TestCheckResourceAttr(
			// 		resourceName, attributeName, "org-col-new-name-bar",
			// 	),
			// },
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
	
	member {
		org_user_id = "%s"
	}

	member {
		org_user_id = "%s"
	}
}
`, testOrganizationID, name, testAccountEmailOrgOwnerInTestOrgUserId, testAccountEmailOrgUserInTestOrgUserId)
}

// func tfProviderOrgOwner(accountEmail string) string {
// 	return fmt.Sprintf(`
// 	provider "bitwarden" {
// 		alias = "org_owner"
// 		master_password = "%s"
// 		server          = "%s"
// 		email           = "%s"

// 		experimental {
// 			embedded_client = true
// 		}
// 	}
// `, testPassword, testServerURL, accountEmail)
// }
