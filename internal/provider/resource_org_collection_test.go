package provider

import (
	"fmt"
	"regexp"
	"strconv"
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
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionNoMembers("org-col-bar"),
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
			// Renaming collection
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionNoMembers("org-col-new-name-bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeName, "org-col-new-name-bar",
					),
					resource.TestMatchResourceAttr(
						resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
					),
				),
			},
			// Importing collection
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: orgCollectionImportID(resourceName),
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccResourceOrgCollectionACLs(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_org_collection.foo_org_col"
	var objectID string

	if useEmbeddedClient {
		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionNoMembers("org-col-bar"),
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
				// 2. Adding one member
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionSingleMember("org-col-bar"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName, schema_definition.AttributeName, "org-col-bar",
						),
						resource.TestMatchResourceAttr(
							resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "1",
						),
						resource.TestCheckTypeSetElemNestedAttrs(
							resourceName, "member.*", map[string]string{
								"id":             testAccountEmailOrgOwnerInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "false",
							},
						),

						getObjectID(resourceName, &objectID),
					),
				},
				// 3. Adding a second member with permission set 1
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionTwoMembers("org-col-bar", false, true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName, schema_definition.AttributeName, "org-col-bar",
						),
						resource.TestMatchResourceAttr(
							resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "2",
						),
						resource.TestCheckTypeSetElemNestedAttrs(
							resourceName, "member.*", map[string]string{
								"id":             testAccountEmailOrgUserInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "true",
							},
						),
					),
				},
				// 4. Changing second member to permissions set 2
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionTwoMembers("org-col-bar", true, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName, schema_definition.AttributeName, "org-col-bar",
						),
						resource.TestMatchResourceAttr(
							resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "2",
						),
						resource.TestCheckTypeSetElemNestedAttrs(
							resourceName, "member.*", map[string]string{
								"id":             testAccountEmailOrgUserInTestOrgUserId,
								"read_only":      "true",
								"hide_passwords": "false",
							},
						),
					),
				},
				// 5. Removing permissions
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionSingleMember("org-col-bar"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName, schema_definition.AttributeName, "org-col-bar",
						),
						resource.TestMatchResourceAttr(
							resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "1",
						),
						resource.TestCheckTypeSetElemNestedAttrs(
							resourceName, "member.*", map[string]string{
								"id":             testAccountEmailOrgOwnerInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "false",
							},
						),
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
	} else {
		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionNoMembers("org-col-bar"),
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
				// Adding one member
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollectionSingleMember("org-col-bar"),
					ExpectError:  regexp.MustCompile("managing collection memberships is only supported by the embedded client"),
				},
			},
		})
	}
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

func tfConfigResourceOrgCollectionTwoMembers(name string, readOnly, hidePasswords bool) string {
	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "%s"
	
	member {
		id = "%s"
		read_only = %s
		hide_passwords = %s
	}

	member {
		id = "%s"
	}
}
`, testOrganizationID, name, testAccountEmailOrgUserInTestOrgUserId, strconv.FormatBool(readOnly), strconv.FormatBool(hidePasswords), testAccountEmailOrgOwnerInTestOrgUserId)
}

func tfConfigResourceOrgCollectionNoMembers(name string) string {
	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "%s"
}
`, testOrganizationID, name)
}

func tfConfigResourceOrgCollectionSingleMember(name string) string {
	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "%s"
	
	member {
		id = "%s"
	}
}
`, testOrganizationID, name, testAccountEmailOrgOwnerInTestOrgUserId)
}
