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
	tfProvider, stop := testOrRealSecretsManagerProvider(t)
	defer stop()

	projectResourceId := "bitwarden_project.foo.id"

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
				Config:      tfProvider + tfConfigResourceSecret("foo", "\"fake-project-id\""),
				ExpectError: regexp.MustCompile("400!=200"),
			},
			// Test Creating Secret with INEXISTENT Project
			{
				Config: tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigResourceSecret("foo", projectResourceId),
				Check:  checkSecret("bitwarden_secret.foo"),
			},
			// Test Sourcing Secret by ID
			{
				Config: tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigResourceSecret("foo", projectResourceId) + tfConfigDataSecretByID("bitwarden_secret.foo.id"),
				Check:  checkSecret("data.bitwarden_secret.foo_data"),
			},
			// Test Sourcing Secret by ID with NO MATCH
			{
				Config:      tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigResourceSecret("foo", projectResourceId) + tfConfigDataSecretByID("\"27a0007a-a517-4f25-8c2e-baf31ca3b034\""),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
			// Test Sourcing Secret by KEY
			{
				Config: tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigResourceSecret("foo", projectResourceId) + tfConfigDataSecretByKey(),
				Check:  checkSecret("data.bitwarden_secret.foo_data"),
			},
			// Test Sourcing Secret with MULTIPLE MATCHES
			{
				Config:      tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigResourceSecret("foo", projectResourceId) + tfConfigResourceSecret("foo2", projectResourceId) + tfConfigDataSecretByKey(),
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
		resource.TestMatchResourceAttr(fullRessourceName, attributeProjectID, regexp.MustCompile("^([a-z0-9-]+)$")),
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

func tfConfigResourceSecret(resourceName, projectResourceId string) string {
	return fmt.Sprintf(`
	resource "bitwarden_secret" "%s" {
		provider = bitwarden

		key = "login-bar"
		value = "value-bar"
		note = "note-bar"
		project_id = %s
	}
`, resourceName, projectResourceId)
}
