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
					conditionalAssertion(useEmbeddedClient,
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "2",
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.0.user_email", testAccountEmailOrgOwner,
						),
						resource.TestMatchResourceAttr(
							resourceName, fmt.Sprintf("member.0.%s", schema_definition.AttributeCollectionMemberOrgMemberId), regexp.MustCompile(regExpId),
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.0.read_only", "false",
						),
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeName, "org-col-bar2",
					),
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
					),
					conditionalAssertion(useEmbeddedClient,
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "2",
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.0.user_email", testAccountEmailOrgOwner,
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.0.read_only", "false",
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.1.user_email", testAccountEmailOrgUser,
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.1.read_only", "false",
						),
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionOneLess("org-col-bar2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeName, "org-col-bar2",
					),
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
					),
					conditionalAssertion(useEmbeddedClient,
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "1",
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.0.user_email", testAccountEmailOrgOwner,
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.0.read_only", "false",
						),
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
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfProviderOrgOwner(testAccountEmailOrgOwner) + tfConfigResourceOrgCollection("org-col-new-name-bar"),
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
	
	member {
		user_email = "%s"
	}

	member {
		user_email = "%s"
	}
}
`, testOrganizationID, name, testAccountEmailOrgOwner, testAccountEmailOrgUser)
}

func tfConfigResourceOrgCollectionOneLess(name string) string {
	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "%s"
	
	member {
		user_email = "%s"
	}
}
`, testOrganizationID, name, testAccountEmailOrgOwner)
}

func tfProviderOrgOwner(accountEmail string) string {
	return fmt.Sprintf(`
	provider "bitwarden" {
		alias = "org_owner"
		master_password = "%s"
		server          = "%s"
		email           = "%s"

		experimental {
			embedded_client = true
		}
	}
`, testPassword, testServerURL, accountEmail)
}
