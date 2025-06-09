//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceItemLoginAttributes(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("reslogin"),
				Check: resource.ComposeTestCheckFunc(
					checkItemLogin(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeNotes, "notes-reslogin",
					),
					getObjectID(resourceName, &objectID),
				),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLogin("resloginmodified"),
				Check: resource.ComposeTestCheckFunc(
					checkItemLogin(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, schema_definition.AttributeNotes, "notes-resloginmodified",
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
	SkipIfOfficialBackend(t, "Creating many items is too slow on the official bitwarden instances.")
	SkipIfOfficialCLI(t, "Creating many items is too slow on the official CLI")

	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemManyLogins(),
			},
		},
	})
}

func TestAccMissingResourceItemLoginIsRecreated(t *testing.T) {
	ensureTestConfigurationReady(t)

	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginSmall(),
				Check: resource.ComposeTestCheckFunc(
					getObjectID("bitwarden_item_login.foo", &objectID),
				),
			},
			{
				Config:             tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginSmall(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginSmall(),
				PreConfig: func() {
					obj := models.Item{ID: objectID, Object: models.ObjectTypeItem}
					err := bwEmbeddedTestClient(t).DeleteItem(t.Context(), obj)
					assert.NoError(t, err)

					if testConfiguration.UseEmbeddedClient {
						return
					}

					// Sync when using the official client, as we removed the object using the API
					// which means the local state is out of sync.
					bwCLITestClient(t).Sync(t.Context())
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
`, testConfiguration.Resources.OrganizationID, testConfiguration.Resources.CollectionID, testConfiguration.Resources.FolderID, source)
}

func checkItemLogin(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		checkItemBase(resourceName),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeFavorite, regexp.MustCompile("^true"),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeLoginUsername, regexp.MustCompile("^test-username$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeLoginPassword, regexp.MustCompile("^test-password$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeLoginTotp, regexp.MustCompile("^1234$"),
		),
		checkItemLoginUriMatches(resourceName),
	)
}

func checkItemLoginUriMatches(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.#", schema_definition.AttributeLoginURIs), regexp.MustCompile("^8$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.0.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^default$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.0.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://default$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.1.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^base_domain$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.1.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://base_domain$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.2.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^host$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.2.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://host$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.3.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^start_with$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.3.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://start_with$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.4.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^exact$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.4.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://exact$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.5.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^regexp$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.5.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://regexp$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.6.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^never$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.6.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://never$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.7.match", schema_definition.AttributeLoginURIs), regexp.MustCompile("^default$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.7.value", schema_definition.AttributeLoginURIs), regexp.MustCompile("^https://default$"),
		),
	)
}
