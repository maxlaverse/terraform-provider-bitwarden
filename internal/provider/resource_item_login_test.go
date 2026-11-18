//go:build integration

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceItemLoginAttributes(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
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

// checkBackendLoginCredentials fetches the item straight from the backend
// (bypassing Terraform state entirely) and asserts its password/username.
// The client is created (and logged in/synced) lazily, when the check
// actually runs, rather than up front when the test steps are constructed --
// otherwise it would sync before the item even exists.
func checkBackendLoginCredentials(t *testing.T, objectID *string, expectedPassword, expectedUsername string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		obj, err := bwEmbeddedTestClient(t).GetItem(t.Context(), models.Item{ID: *objectID, Object: models.ObjectTypeItem})
		if err != nil {
			return err
		}
		if obj.Login.Password != expectedPassword {
			return fmt.Errorf("expected backend login password %q, got %q", expectedPassword, obj.Login.Password)
		}
		if obj.Login.Username != expectedUsername {
			return fmt.Errorf("expected backend login username %q, got %q", expectedUsername, obj.Login.Username)
		}
		return nil
	}
}

// TestAccResourceItemLoginWriteOnly covers the maintainer-requested scenarios
// for password_wo/username_wo/wo_version:
//   - bumping wo_version alongside new password_wo/username_wo values updates
//     the item, even though neither is ever in state;
//   - the plan stays empty when wo_version is unchanged, even if
//     password_wo/username_wo would evaluate to a different value on this
//     apply (e.g. fed by an ephemeral resource that produces a fresh value
//     every time);
//   - changing an unrelated attribute while wo_version is unchanged must not
//     change the vault password/username, even though password_wo/username_wo
//     are reconfigured with different (would-be) values alongside it.
func TestAccResourceItemLoginWriteOnly(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo")
	var objectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-1", "wo-username-1", 1, ""),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginUsername),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginWOVersion, "1"),
					checkBackendLoginCredentials(t, &objectID, "wo-secret-1", "wo-username-1"),
				),
			},
			// Bumping wo_version alongside new password_wo/username_wo values must
			// update the item, even though neither is ever in state.
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-2", "wo-username-2", 2, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, schema_definition.AttributeID, &objectID),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginUsername),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginWOVersion, "2"),
					checkBackendLoginCredentials(t, &objectID, "wo-secret-2", "wo-username-2"),
				),
			},
			// Same config, same wo_version: the plan must stay empty.
			{
				Config:             tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-2", "wo-username-2", 2, ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// wo_version unchanged, but password_wo/username_wo now evaluate to a
			// different literal (simulating an ephemeral resource producing a new
			// value on every apply): the plan must still be empty, since
			// write-only attributes never contribute to the diff.
			{
				Config:             tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-3-unused", "wo-username-3-unused", 2, ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// wo_version unchanged, but an unrelated attribute (totp) changes, and
			// password_wo/username_wo evaluate to different (would-be) values:
			// totp must update, but the vault password/username must not change.
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-4-unused", "wo-username-4-unused", 2, "4321"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginTotp, "4321"),
					checkBackendLoginCredentials(t, &objectID, "wo-secret-2", "wo-username-2"),
				),
			},
		},
	})
}

// TestAccResourceItemLoginWriteOnlyPartial checks that password_wo can be
// paired with a plain (non-write-only) username, and the other way around --
// the two write-only attributes must not be required together.
func TestAccResourceItemLoginWriteOnlyPartial(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo-partial")
	var objectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name        	= %q
		username    	= "plain-username"
		password_wo 	= "wo-secret-1"
		wo_version  	= 1
	}
`, testConfiguration.Resources.OrganizationID, name),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginUsername, "plain-username"),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					checkBackendLoginCredentials(t, &objectID, "wo-secret-1", "plain-username"),
				),
			},
			// Bumping wo_version rotates the password alone; the plain username is
			// left as configured, without ever needing username_wo.
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name        	= %q
		username    	= "plain-username"
		password_wo 	= "wo-secret-2"
		wo_version  	= 2
	}
`, testConfiguration.Resources.OrganizationID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginUsername, "plain-username"),
					checkBackendLoginCredentials(t, &objectID, "wo-secret-2", "plain-username"),
				),
			},
		},
	})
}

// TestAccResourceItemLoginMigrateToWriteOnly checks that migrating an item
// from a plain password to password_wo removes the secret from state.
func TestAccResourceItemLoginMigrateToWriteOnly(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo-migrate")
	var objectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name     		= %q
		username 		= "plain-username"
		password 		= "plain-password"
	}
`, testConfiguration.Resources.OrganizationID, name),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginPassword, "plain-password"),
				),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "migrated-secret", "migrated-username", 1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginUsername),
					checkBackendLoginCredentials(t, &objectID, "migrated-secret", "migrated-username"),
				),
			},
		},
	})
}

func tfConfigResourceItemLoginWO(name, password, username string, version int, totp string) string {
	totpAttr := ""
	if totp != "" {
		totpAttr = fmt.Sprintf("\t\ttotp        = %q\n", totp)
	}
	return fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name        	= %q
		password_wo 	= %q
		username_wo 	= %q
		wo_version  	= %d
%s	}
`, testConfiguration.Resources.OrganizationID, name, password, username, version, totpAttr)
}

func TestAccResourceItemLoginAddRemoveCollection(t *testing.T) {
	t.Skip("Skipping test until we figure out how sharing work with official backend")
	SkipIfNonPremiumTestAccount(t, "Having more than one collection is not supported with non-premium test accounts")

	ensureTestConfigurationReady(t)

	// Create second collection via embedded client (supports membership).
	// Only for premium or non-official accounts. Keeps the test focused on item collection_ids add/remove.
	bwClient := bwEmbeddedTestClient(t).(embedded.PasswordManagerClient)
	loginTestCol, err := bwClient.CreateOrganizationCollection(t.Context(), models.OrgCollection{
		Object:         models.ObjectTypeOrgCollection,
		OrganizationID: testConfiguration.Resources.OrganizationID,
		Name:           testName("login-test-col"),
		Users: []models.OrgCollectionMember{
			{Id: testConfiguration.Accounts[testAccountOrgOwner].UserIdInTestOrganization, Manage: true},
		},
		Groups: []models.OrgCollectionMember{},
	})
	if err != nil {
		t.Fatalf("creating login test collection: %v", err)
	}

	loginTestCollectionID := loginTestCol.ID
	t.Cleanup(func() {
		if err := bwClient.DeleteOrganizationCollection(context.Background(), models.OrgCollection{
			Object:         models.ObjectTypeOrgCollection,
			ID:             loginTestCollectionID,
			OrganizationID: testConfiguration.Resources.OrganizationID,
		}); err != nil {
			t.Logf("cleaning up login test collection %s: %v", loginTestCollectionID, err)
		}
	})

	resourceName := "bitwarden_item_login.foo"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWithCollections("reslogin", []string{testConfiguration.Resources.CollectionID}),
				Check: resource.ComposeTestCheckFunc(
					checkItemLogin(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, fmt.Sprintf("%s.#", schema_definition.AttributeCollectionIDs), "1",
					),
					resource.TestCheckResourceAttr(
						resourceName, fmt.Sprintf("%s.0", schema_definition.AttributeCollectionIDs), testConfiguration.Resources.CollectionID,
					),
				),
			},
			// Add second collection (2 collections)
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWithCollections("reslogin", []string{testConfiguration.Resources.CollectionID, loginTestCollectionID}),
				Check: resource.ComposeTestCheckFunc(
					checkResourceAttrListEqualsSet(resourceName, schema_definition.AttributeCollectionIDs, []*string{&testConfiguration.Resources.CollectionID, &loginTestCollectionID}),
				),
			},
			// Remove second collection (back to 1)
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWithCollections("reslogin", []string{testConfiguration.Resources.CollectionID}),
				Check: resource.ComposeTestCheckFunc(
					checkItemLogin(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, fmt.Sprintf("%s.#", schema_definition.AttributeCollectionIDs), "1",
					),
					resource.TestCheckTypeSetElemAttr(
						resourceName, schema_definition.AttributeCollectionIDs+".*", testConfiguration.Resources.CollectionID,
					),
				),
			},
		},
	})
}

func TestAccResourceItemLoginMany(t *testing.T) {
	SkipIfOfficialBackend(t, "Creating many items is too slow on the official bitwarden instances.")
	SkipIfOfficialCLI(t, "Creating many items is too slow on the official CLI")

	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
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
		ProtoV6ProviderFactories: providerFactories,
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
	return fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		name     			= "%s"
	}
`, testName("login-bar"))
}

func tfConfigResourceItemManyLogins() string {
	return fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		count    = 100
		name     = "%s Login Item ${count.index + 1}"
	}
`, testName("many"))
}

func tfConfigResourceItemLogin(source string) string {
	return tfConfigResourceItemLoginWithCollections(source, []string{testConfiguration.Resources.CollectionID})
}

const tfConfigResourceItemLoginTemplate = `
	resource "bitwarden_item_login" "foo" {
		provider 			= bitwarden

		organization_id     = "%s"
		collection_ids		= %s
		folder_id 			= "%s"
		username 			= "%s"
		password 			= "test-password"
		totp 				= "1234"
		name     			= "%s"
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
`

func tfConfigResourceItemLoginWithCollections(source string, collectionIDs []string) string {
	collectionIDsHCL := "[]"
	if len(collectionIDs) > 0 {
		quoted := make([]string, len(collectionIDs))
		for i, id := range collectionIDs {
			quoted[i] = fmt.Sprintf("%q", id)
		}
		collectionIDsHCL = "[" + strings.Join(quoted, ", ") + "]"
	}
	return fmt.Sprintf(tfConfigResourceItemLoginTemplate,
		testConfiguration.Resources.OrganizationID, collectionIDsHCL, testConfiguration.Resources.FolderID, testUsername(), testName("login-bar"), source)
}

// tfConfigResourceItemLoginWithCollectionRefs is like tfConfigResourceItemLoginWithCollections but allows Terraform refs in collection_ids (e.g. bitwarden_org_collection.foo.id).
func tfConfigResourceItemLoginWithCollectionRefs(source string, literalIDs []string, refs []string) string {
	parts := make([]string, 0, len(literalIDs)+len(refs))
	for _, id := range literalIDs {
		parts = append(parts, fmt.Sprintf("%q", id))
	}
	parts = append(parts, refs...)
	collectionIDsHCL := "[" + strings.Join(parts, ", ") + "]"
	return fmt.Sprintf(tfConfigResourceItemLoginTemplate,
		testConfiguration.Resources.OrganizationID, collectionIDsHCL, testConfiguration.Resources.FolderID, testUsername(), testName("login-bar"), source)
}

func checkItemLogin(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		checkItemBase(resourceName),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeFavorite, regexp.MustCompile("^true"),
		),
		resource.TestCheckResourceAttr(
			resourceName, schema_definition.AttributeLoginUsername, testUsername(),
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
