//go:build integration

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

// backendLoginItem fetches the item straight from the backend, bypassing
// Terraform state entirely. A fresh client is logged in (and synced) on every
// call, so it sees what the last apply wrote. Errors are returned rather than
// raised with t.Fatal, which would abort the test before its destroy step.
func backendLoginItem(t *testing.T, objectID string) (*models.Item, error) {
	client := embedded.NewPasswordManagerClient(testConfiguration.ServerURL, testDeviceIdentifer, testDeviceVersion)
	acc := testConfiguration.Accounts[testAccountFullAdmin]
	if err := client.LoginWithAPIKey(t.Context(), acc.Password, acc.ClientID, acc.ClientSecret); err != nil {
		return nil, err
	}
	return client.GetItem(t.Context(), models.Item{ID: objectID, Object: models.ObjectTypeItem})
}

// checkBackendLoginCredentials asserts the password/username stored in the
// backend. It is lazy on purpose: building the client when the test steps are
// constructed would sync before the item even exists.
func checkBackendLoginCredentials(t *testing.T, objectID *string, expectedPassword, expectedUsername string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		obj, err := backendLoginItem(t, *objectID)
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

// captureBackendLoginPassword stores the password currently in the backend.
func captureBackendLoginPassword(t *testing.T, objectID *string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		obj, err := backendLoginItem(t, *objectID)
		if err != nil {
			return err
		}
		*target = obj.Login.Password
		return nil
	}
}

// checkBackendLoginPassword compares the backend password with a previously
// captured one: equal when wantSame, different otherwise.
func checkBackendLoginPassword(t *testing.T, objectID *string, captured *string, wantSame bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		obj, err := backendLoginItem(t, *objectID)
		if err != nil {
			return err
		}
		if (obj.Login.Password == *captured) != wantSame {
			return fmt.Errorf("backend password unchanged=%t, want unchanged=%t", obj.Login.Password == *captured, wantSame)
		}
		return nil
	}
}

// TestAccResourceItemLoginWriteOnly covers the maintainer-requested scenarios
// for password_wo/username_wo/wo_version:
//   - bumping wo_version alongside new password_wo/username_wo values updates
//     the item, even though neither is ever in state;
//   - refreshing never copies the secret back into state;
//   - the plan is empty when nothing changed;
//   - changing an unrelated attribute while wo_version is unchanged must not
//     change the vault password/username, even though password_wo/username_wo
//     are reconfigured with different (would-be) values alongside it.
//
// TestAccResourceItemLoginWriteOnlyEphemeralSource covers the same with a real
// ephemeral resource producing a new value on every apply.
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
			// Refreshing must not copy the secret from the backend into state.
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginUsername),
					checkBackendLoginCredentials(t, &objectID, "wo-secret-2", "wo-username-2"),
				),
			},
			// Same config, same wo_version: nothing to apply. Checked as a plan
			// check on a real apply step so that a password/username planned as
			// "known after apply" would fail it.
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-2", "wo-username-2", 2, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
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

// TestAccResourceItemLoginWriteOnlyPartialReverse is the other half of
// TestAccResourceItemLoginWriteOnlyPartial: a plain password with username_wo.
func TestAccResourceItemLoginWriteOnlyPartialReverse(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo-partial-reverse")
	var objectID string

	config := func(username string, version int) string {
		return tfConfigPasswordManagerProvider(testAccountFullAdmin) + fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name        	= %q
		password    	= "plain-password"
		username_wo 	= %q
		wo_version  	= %d
	}
`, testConfiguration.Resources.OrganizationID, name, username, version)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config("wo-username-1", 1),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginPassword, "plain-password"),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginUsername),
					checkBackendLoginCredentials(t, &objectID, "plain-password", "wo-username-1"),
				),
			},
			{
				Config: config("wo-username-2", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginPassword, "plain-password"),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginUsername),
					checkBackendLoginCredentials(t, &objectID, "plain-password", "wo-username-2"),
				),
			},
		},
	})
}

// TestAccResourceItemLoginMigrateToWriteOnlySameVersion migrates from a plain
// password to password_wo while wo_version stays the same. The change must
// still be planned and applied, and the secret must leave the state.
func TestAccResourceItemLoginMigrateToWriteOnlySameVersion(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo-migrate-same-version")
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
		password    	= "plain-password"
		wo_version  	= 1
	}
`, testConfiguration.Resources.OrganizationID, name),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginPassword, "plain-password"),
				),
			},
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + fmt.Sprintf(`
	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name        	= %q
		username    	= "plain-username"
		password_wo 	= "migrated-secret"
		wo_version  	= 1
	}
`, testConfiguration.Resources.OrganizationID, name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, schema_definition.AttributeID, &objectID),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					checkBackendLoginCredentials(t, &objectID, "migrated-secret", "plain-username"),
				),
			},
		},
	})
}

// TestAccResourceItemLoginWriteOnlyImport checks that an item created with
// write-only attributes can be imported. The imported state mirrors the
// backend, so it holds the plaintext values and no wo_version: that is
// expected, and why password, username and wo_version are ignored below.
func TestAccResourceItemLoginWriteOnlyImport(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo-import")
	var objectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemLoginWO(name, "wo-secret-1", "wo-username-1", 1, ""),
				Check:  getObjectID(resourceName, &objectID),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       func(*terraform.State) (string, error) { return objectID, nil },
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{schema_definition.AttributeLoginPassword, schema_definition.AttributeLoginUsername, schema_definition.AttributeLoginWOVersion},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}
					if got := states[0].Attributes[schema_definition.AttributeLoginPassword]; got != "wo-secret-1" {
						return fmt.Errorf("expected imported password to mirror the backend, got %q", got)
					}
					return nil
				},
			},
		},
	})
}

// TestAccResourceItemLoginWriteOnlyEphemeralSource feeds password_wo from a
// real ephemeral resource, which returns a different value on every plan and
// apply, as the maintainer's review describes. Only wo_version may decide
// whether the vault password changes.
func TestAccResourceItemLoginWriteOnlyEphemeralSource(t *testing.T) {
	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_login.foo"
	name := testName("login-wo-ephemeral")
	var objectID, initialPassword string

	config := func(version int, totp string) string {
		totpAttr := ""
		if totp != "" {
			totpAttr = fmt.Sprintf("totp = %q", totp)
		}
		return tfConfigPasswordManagerProvider(testAccountFullAdmin) + fmt.Sprintf(`
	ephemeral "random_password" "generated" {
		length = 24
	}

	resource "bitwarden_item_login" "foo" {
		provider 		= bitwarden

		organization_id = %q
		name        	= %q
		username    	= "ephemeral-username"
		password_wo 	= ephemeral.random_password.generated.result
		wo_version  	= %d
		%s
	}
`, testConfiguration.Resources.OrganizationID, name, version, totpAttr)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {Source: "hashicorp/random"},
		},
		Steps: []resource.TestStep{
			{
				Config: config(1, ""),
				Check: resource.ComposeTestCheckFunc(
					getObjectID(resourceName, &objectID),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					captureBackendLoginPassword(t, &objectID, &initialPassword),
				),
			},
			// Same wo_version: the ephemeral value is different, the plan is empty
			// and the vault password is untouched.
			{
				Config: config(1, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: checkBackendLoginPassword(t, &objectID, &initialPassword, true),
			},
			// Another attribute changes: still the same vault password.
			{
				Config: config(1, "4321"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeLoginTotp, "4321"),
					checkBackendLoginPassword(t, &objectID, &initialPassword, true),
				),
			},
			// Bumping wo_version applies the new ephemeral value.
			{
				Config: config(2, "4321"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, schema_definition.AttributeID, &objectID),
					resource.TestCheckNoResourceAttr(resourceName, schema_definition.AttributeLoginPassword),
					checkBackendLoginPassword(t, &objectID, &initialPassword, false),
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
