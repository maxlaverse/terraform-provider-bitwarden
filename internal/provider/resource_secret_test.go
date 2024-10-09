package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceSecret(t *testing.T) {
	if len(testProjectId) == 0 {
		t.Skip("Skipping test due to missing project_id")
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigSecretsManagerProvider() + tfConfigResourceSecret(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("bitwarden_secret.foo", attributeID, regexp.MustCompile("^([a-z0-9-]+)$")),
					resource.TestCheckResourceAttr("bitwarden_secret.foo", attributeKey, "login-bar"),
					resource.TestCheckResourceAttr("bitwarden_secret.foo", attributeValue, "value-bar"),
					resource.TestCheckResourceAttr("bitwarden_secret.foo", attributeNote, "note-bar"),
					resource.TestCheckResourceAttr("bitwarden_secret.foo", attributeProjectID, testProjectId),
				),
			},
			{
				Config: tfConfigSecretsManagerProvider() + tfConfigResourceSecret() + tfConfigDataSecret(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.bitwarden_secret.foo_data", attributeID, regexp.MustCompile("^([a-z0-9-]+)$")),
					resource.TestCheckResourceAttr("data.bitwarden_secret.foo_data", attributeKey, "login-bar"),
					resource.TestCheckResourceAttr("data.bitwarden_secret.foo_data", attributeValue, "value-bar"),
					resource.TestCheckResourceAttr("data.bitwarden_secret.foo_data", attributeNote, "note-bar"),
					resource.TestCheckResourceAttr("data.bitwarden_secret.foo_data", attributeProjectID, testProjectId),
				),
			},
		},
	})
}

func tfConfigDataSecret() string {
	return `
data "bitwarden_secret" "foo_data" {
	provider	= bitwarden

	id 	= bitwarden_secret.foo.id
}
`
}

func tfConfigResourceSecret() string {
	return fmt.Sprintf(`
	resource "bitwarden_secret" "foo" {
		provider = bitwarden

		key = "login-bar"
		value = "value-bar"
		note = "note-bar"
		project_id ="%s"
	}
`, testProjectId)
}
