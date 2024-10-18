package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceSecretSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfConfigSecretsManagerProvider() + tfConfigDataSecretWithoutAnyInput(),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      tfConfigSecretsManagerProvider() + tfConfigDataSecretTooManyInput(),
				ExpectError: regexp.MustCompile(": conflicts"),
			},
		},
	})
}

func TestResourceSecret(t *testing.T) {
	tfConfigSecretsManagerProvider()
	if len(testProjectId) == 0 {
		t.Skip("Skipping test due to missing project_id")
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfConfigSecretsManagerProvider() + tfConfigDataSecretWithoutAnyInput(),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      tfConfigSecretsManagerProvider() + tfConfigDataSecretTooManyInput(),
				ExpectError: regexp.MustCompile(": conflicts"),
			},
			{
				Config: tfConfigSecretsManagerProvider() + tfConfigResourceSecret("foo"),
				Check:  checkSecret("bitwarden_secret.foo"),
			},
			// Test Sourcing Secret by ID
			{
				Config: tfConfigSecretsManagerProvider() + tfConfigResourceSecret("foo") + tfConfigDataSecretByID("bitwarden_secret.foo.id"),
				Check:  checkSecret("data.bitwarden_secret.foo_data"),
			},
			// Test Sourcing Secret by ID with NO MATCH
			{
				Config:      tfConfigSecretsManagerProvider() + tfConfigResourceSecret("foo") + tfConfigDataSecretByID("\"27a0007a-a517-4f25-8c2e-baf31ca3b034\""),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
			// Test Sourcing Secret by KEY
			{
				Config: tfConfigSecretsManagerProvider() + tfConfigResourceSecret("foo") + tfConfigDataSecretByKey(),
				Check:  checkSecret("data.bitwarden_secret.foo_data"),
			},
			// Test Sourcing Secret with MULTIPLE MATCHES
			{
				Config:      tfConfigSecretsManagerProvider() + tfConfigResourceSecret("foo") + tfConfigResourceSecret("foo2") + tfConfigDataSecretByKey(),
				ExpectError: regexp.MustCompile("Error: too many objects found"),
			},
		},
	})
}

func checkSecret(fullRessourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(fullRessourceName, attributeID, regexp.MustCompile("^([a-z0-9-]+)$")),
		resource.TestCheckResourceAttr(fullRessourceName, attributeKey, "login-bar"),
		resource.TestCheckResourceAttr(fullRessourceName, attributeValue, "value-bar"),
		resource.TestCheckResourceAttr(fullRessourceName, attributeNote, "note-bar"),
		resource.TestCheckResourceAttr(fullRessourceName, attributeProjectID, testProjectId),
	)
}
func tfConfigDataSecretByID(id string) string {
	return fmt.Sprintf(`
data "bitwarden_secret" "foo_data" {
	provider	= bitwarden

	id 	= %s
}
`, id)
}

func tfConfigDataSecretByKey() string {
	return `
data "bitwarden_secret" "foo_data" {
	provider	= bitwarden

	key = "login-bar"
}
`
}

func tfConfigDataSecretWithoutAnyInput() string {
	return `
data "bitwarden_secret" "foo_data" {
	provider	= bitwarden
}
`
}

func tfConfigDataSecretTooManyInput() string {
	return `
data "bitwarden_secret" "foo_data" {
	provider	= bitwarden

	key = "something"
	id = "something"
}
`
}

func tfConfigResourceSecret(resourceName string) string {
	return fmt.Sprintf(`
	resource "bitwarden_secret" "%s" {
		provider = bitwarden

		key = "login-bar"
		value = "value-bar"
		note = "note-bar"
		project_id ="%s"
	}
`, resourceName, testProjectId)
}
