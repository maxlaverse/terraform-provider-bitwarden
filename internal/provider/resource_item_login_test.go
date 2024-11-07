package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceItemLoginAttributes(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resourceName := "bitwarden_item_login.foo"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemLogin("reslogin"),
				Check: resource.ComposeTestCheckFunc(
					checkItemLogin(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, attributeNotes, "notes-reslogin",
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemLogin("resloginmodified"),
				Check: resource.ComposeTestCheckFunc(
					checkItemLogin(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, attributeNotes, "notes-resloginmodified",
					),
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

func TestAccResourceItemLoginMany(t *testing.T) {
	if !useEmbeddedClient {
		t.Skip("Skipping test because using the official client to create many items is too slow")
	}
	ensureVaultwardenConfigured(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemManyLogins(),
			},
		},
	})
}

func TestAccMissingResourceItemLoginIsRecreated(t *testing.T) {
	ensureVaultwardenConfigured(t)

	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemLoginSmall(),
				Check: resource.ComposeTestCheckFunc(
					getObjectID("bitwarden_item_login.foo", &objectID),
				),
			},
			{
				Config:             tfConfigPasswordManagerProvider() + tfConfigResourceItemLoginSmall(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: tfConfigPasswordManagerProvider() + tfConfigResourceItemLoginSmall(),
				PreConfig: func() {
					obj := models.Object{ID: objectID, Object: models.ObjectTypeItem}
					err := bwTestClient(t).DeleteObject(context.Background(), obj)
					assert.NoError(t, err)

					if useEmbeddedClient {
						return
					}

					// Sync when using the official client, as we removed the object using the API
					// which means the local state is out of sync.
					bwOfficialTestClient(t).Sync(context.Background())
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func tfConfigResourceItemLoginSmall() string {
	return `
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		name     			= "login-bar"
	}
`
}

func tfConfigResourceItemManyLogins() string {
	return `
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		count    = 100
		name     = "Login Item ${count.index + 1}"
	}
`
}

func tfConfigResourceItemLogin(source string) string {
	return fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		organization_id     = "%s"
		collection_ids		= ["%s"]
		folder_id 			= "%s"
		username 			= "test-username"
		password 			= "test-password"
		totp 				= "1234"
		name     			= "login-bar"
		notes 				= "notes-%s"
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

		uri {
			match = "default"
			value = "https://default"
		}

		uri {
			match = "base_domain"
			value = "https://base_domain"
		}

		uri {
			match = "host"
			value = "https://host"
		}

		uri {
			match = "start_with"
			value = "https://start_with"
		}

		uri {
			match = "exact"
			value = "https://exact"
		}

		uri {
			match = "regexp"
			value = "https://regexp"
		}

		uri {
			match = "never"
			value = "https://never"
		}

		uri {
			value = "https://default"
		}
	}
`, testOrganizationID, testCollectionID, testFolderID, source)
}

func checkItemLogin(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		checkItemGeneral(resourceName),
		resource.TestMatchResourceAttr(
			resourceName, attributeLoginUsername, regexp.MustCompile("^test-username$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeLoginPassword, regexp.MustCompile("^test-password$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeLoginTotp, regexp.MustCompile("^1234$"),
		),
		checkItemLoginUriMatches(resourceName),
	)
}

func checkItemLoginUriMatches(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.#", attributeLoginURIs), regexp.MustCompile("^8$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.0.match", attributeLoginURIs), regexp.MustCompile("^default$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.0.value", attributeLoginURIs), regexp.MustCompile("^https://default$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.1.match", attributeLoginURIs), regexp.MustCompile("^base_domain$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.1.value", attributeLoginURIs), regexp.MustCompile("^https://base_domain$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.2.match", attributeLoginURIs), regexp.MustCompile("^host$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.2.value", attributeLoginURIs), regexp.MustCompile("^https://host$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.3.match", attributeLoginURIs), regexp.MustCompile("^start_with$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.3.value", attributeLoginURIs), regexp.MustCompile("^https://start_with$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.4.match", attributeLoginURIs), regexp.MustCompile("^exact$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.4.value", attributeLoginURIs), regexp.MustCompile("^https://exact$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.5.match", attributeLoginURIs), regexp.MustCompile("^regexp$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.5.value", attributeLoginURIs), regexp.MustCompile("^https://regexp$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.6.match", attributeLoginURIs), regexp.MustCompile("^never$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.6.value", attributeLoginURIs), regexp.MustCompile("^https://never$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.7.match", attributeLoginURIs), regexp.MustCompile("^default$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.7.value", attributeLoginURIs), regexp.MustCompile("^https://default$"),
		),
	)
}
