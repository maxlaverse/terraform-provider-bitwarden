package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceSecretSchema(t *testing.T) {
	tfProvider, _ := tfConfigSecretsManagerProvider()

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfProvider + tfConfigDataSecretWithoutAnyInput(),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      tfProvider + tfConfigDataSecretTooManyInput(),
				ExpectError: regexp.MustCompile(": conflicts"),
			},
		},
	})
}

func TestResourceSecret(t *testing.T) {
	tfProvider, testProjectId, stop := testOrRealSecretsManagerProvider(t)
	defer stop()

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfProvider + tfConfigDataSecretWithoutAnyInput(),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      tfProvider + tfConfigDataSecretTooManyInput(),
				ExpectError: regexp.MustCompile(": conflicts"),
			},
			{
				Config: tfProvider + tfConfigResourceSecret("foo", testProjectId),
				Check:  checkSecret("bitwarden_secret.foo", testProjectId),
			},
			// Test Sourcing Secret by ID
			{
				Config: tfProvider + tfConfigResourceSecret("foo", testProjectId) + tfConfigDataSecretByID("bitwarden_secret.foo.id"),
				Check:  checkSecret("data.bitwarden_secret.foo_data", testProjectId),
			},
			// Test Sourcing Secret by ID with NO MATCH
			{
				Config:      tfProvider + tfConfigResourceSecret("foo", testProjectId) + tfConfigDataSecretByID("\"27a0007a-a517-4f25-8c2e-baf31ca3b034\""),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
			// Test Sourcing Secret by KEY
			{
				Config: tfProvider + tfConfigResourceSecret("foo", testProjectId) + tfConfigDataSecretByKey(),
				Check:  checkSecret("data.bitwarden_secret.foo_data", testProjectId),
			},
			// Test Sourcing Secret with MULTIPLE MATCHES
			{
				Config:      tfProvider + tfConfigResourceSecret("foo", testProjectId) + tfConfigResourceSecret("foo2", testProjectId) + tfConfigDataSecretByKey(),
				ExpectError: regexp.MustCompile("Error: too many objects found"),
			},
		},
	})
}

func checkSecret(fullRessourceName, testProjectId string) resource.TestCheckFunc {
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

func tfConfigResourceSecret(resourceName, testProjectId string) string {
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
