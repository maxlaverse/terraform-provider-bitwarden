package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceItemLogin(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfTestProvider() + tfTestResourceItemLogin(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeFolderID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeLoginPassword, regexp.MustCompile("^test-password$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeLoginTotp, regexp.MustCompile("^1234$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeLoginUsername, regexp.MustCompile("^test-username$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeName, regexp.MustCompile("^bar$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeNotes, regexp.MustCompile("^notes$"),
					),
					resource.TestMatchResourceAttr(
						"bitwarden_item_login.foo", attributeOrganizationID, regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"),
					),
				),
			},
			{
				Config:           tfTestProvider() + `resource "bitwarden_item_login" "foo_import" { provider = bitwarden }`,
				ResourceName:     "bitwarden_item_login.foo_import",
				ImportState:      true,
				ImportStateId:    testItemLoginID,
				ImportStateCheck: hasOneInstanceState,
			},
		},
	})
}

func tfTestResourceItemLogin() string {
	return fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		organization_id		= "%s"
		collection_ids		= ["%s"]
		folder_id 			= "%s"
		username 			= "test-username"
		password 			= "test-password"
		totp 				= "1234"
		name     			= "bar"
		notes 				= "notes"
	}
`, testOrganizationID, testCollectionID, testFolderID)
}
