//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccResourceItemSecureNote(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_item_secure_note.foo"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemSecureNote(),
				Check: resource.ComposeTestCheckFunc(
					checkItemSecureNote(resourceName),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     objectID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func tfConfigResourceItemSecureNote() string {
	return fmt.Sprintf(`
	resource "bitwarden_item_secure_note" "foo" {
		provider 			= bitwarden

		organization_id     = "%s"
		collection_ids		= ["%s"]
		folder_id 			= "%s"
		name     			= "secure-bar"
		notes 				= "notes"
		reprompt			= true
		favorite            = true 

		field {
			name = "field-text"
			text = "value-text"
		}

		field {
			name    = "field-boolean"
			boolean = true
		}

		field {
			name   = "field-hidden"
			hidden = "value-hidden"
		}
	}
`, testOrganizationID, testCollectionID, testFolderID)
}

func checkItemSecureNote(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeFavorite, regexp.MustCompile("^true"),
		),
	)
}
