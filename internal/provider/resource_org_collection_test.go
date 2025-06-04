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

func TestAccResourceOrgCollection(t *testing.T) {
	SkipIfOfficialBackend(t, "Bitwarden has stopped accepting the creation of collections without a member with manage permissions")

	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_org_collection.foo_org_col"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false),
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
				Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-new-name-bar", false),
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

	if IsOfficialBackend() {
		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				// 1. Create an org-collection with ourself as the only member
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", true),
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
								"manage":         "true",
							},
						),

						getObjectID(resourceName, &objectID),
					),
				},
				// 2. Add another member to the collection
				{
					ResourceName: resourceName,
					Config: tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", true,
						memberBlock(testAccountEmailOrgUserInTestOrgUserId, nil),
					),
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
								"id":             testAccountEmailOrgOwnerInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "false",
								"manage":         "true",
							},
						),
						resource.TestCheckTypeSetElemNestedAttrs(
							resourceName, "member.*", map[string]string{
								"id":             testAccountEmailOrgUserInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "false",
							},
						),

						getObjectID(resourceName, &objectID),
					),
				},
				// 3. Remove the other member from the collection
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName, schema_definition.AttributeName, "org-col-bar",
						),
						resource.TestCheckResourceAttr(
							resourceName, "member.#", "1",
						),
						resource.TestCheckTypeSetElemNestedAttrs(
							resourceName, "member.*", map[string]string{
								"id":             testAccountEmailOrgOwnerInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "false",
								"manage":         "true",
							},
						),
					),
				},
				// 4. Import the collection
				{
					ResourceName:      resourceName,
					ImportStateIdFunc: orgCollectionImportID(resourceName),
					ImportState:       true,
					ImportStateVerify: false,
				},
			},
		})
	} else if IsVaultwardenBackend() && useEmbeddedClient {
		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					ResourceName: resourceName,
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false),
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
					Config: tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false,
						memberBlock(testAccountEmailOrgUserInTestOrgUserId, nil),
					),
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
								"id":             testAccountEmailOrgUserInTestOrgUserId,
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
					Config: tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false,
						memberBlock(testAccountEmailOrgUserInTestOrgUserId, nil),
						memberBlock(testAccountEmailOrgManagerInTestOrgUserId, map[string]string{
							"read_only":      "false",
							"hide_passwords": "true",
						}),
					),
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
								"id":             testAccountEmailOrgManagerInTestOrgUserId,
								"read_only":      "false",
								"hide_passwords": "true",
								"manage":         "false",
							},
						),
					),
				},
				// 4. Changing second member to permissions set 2
				{
					ResourceName: resourceName,
					Config: tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false,
						memberBlock(testAccountEmailOrgUserInTestOrgUserId, nil),
						memberBlock(testAccountEmailOrgManagerInTestOrgUserId, map[string]string{
							"read_only":      "true",
							"hide_passwords": "false",
						}),
					),
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
								"id":             testAccountEmailOrgManagerInTestOrgUserId,
								"read_only":      "true",
								"hide_passwords": "false",
							},
						),
					),
				},
				// 5. Removing permissions
				{
					ResourceName: resourceName,
					Config: tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false,
						memberBlock(testAccountEmailOrgUserInTestOrgUserId, nil),
					),
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
								"id":             testAccountEmailOrgUserInTestOrgUserId,
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
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", false),
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
					Config:       tfConfigPasswordManagerProvider() + tfConfigResourceOrgCollection("org-col-bar", true),
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

func tfConfigResourceOrgCollection(name string, includeOurselves bool, members ...string) string {
	var allMembers []string
	if includeOurselves {
		allMembers = append(allMembers, memberBlock(testAccountEmailOrgOwnerInTestOrgUserId, map[string]string{
			"manage": "true",
		}))
	}
	allMembers = append(allMembers, members...)

	return fmt.Sprintf(`
	resource "bitwarden_org_collection" "foo_org_col" {
	provider	= bitwarden

	organization_id = "%s"

	name     = "%s"
	%s
}
`, testOrganizationID, name, strings.Join(allMembers, "\n\t"))
}

func memberBlock(id string, attrs map[string]string) string {
	var block strings.Builder
	block.WriteString(fmt.Sprintf("\n\tmember {\n\t\tid = \"%s\"", id))
	for k, v := range attrs {
		block.WriteString(fmt.Sprintf("\n\t\t%s = %s", k, v))
	}
	block.WriteString("\n\t}")
	return block.String()
}
