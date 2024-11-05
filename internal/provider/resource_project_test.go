package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceProject(t *testing.T) {
	tfProvider, stop := testOrRealSecretsManagerProvider(t)
	defer stop()

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfProvider + tfConfigResourceProject("foo", "project-foo"),
				Check:  checkProject("bitwarden_project.foo", "project-foo"),
			},
			// Test Sourcing Project by ID
			{
				Config: tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigDataProjectByID("bitwarden_project.foo.id"),
				Check:  checkProject("data.bitwarden_project.foo_data", "project-foo"),
			},
			// Test Sourcing Project by ID with NO MATCH
			{
				Config:      tfProvider + tfConfigResourceProject("foo", "project-foo") + tfConfigDataProjectByID("\"27a0007a-a517-4f25-8c2e-baf31ca3b034\""),
				ExpectError: regexp.MustCompile("Error: object not found"),
			},
			// Test Editing Project
			{
				Config: tfProvider + tfConfigResourceProject("foo", "project-bar"),
				Check:  checkProject("bitwarden_project.foo", "project-bar"),
			},
		},
	})
}

func checkProject(fullRessourceName, projectName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(fullRessourceName, attributeID, regexp.MustCompile("^([a-z0-9-]+)$")),
		resource.TestCheckResourceAttr(fullRessourceName, attributeName, projectName),
		resource.TestMatchResourceAttr(fullRessourceName, attributeOrganizationID, regexp.MustCompile("^([a-z0-9-]+)$")),
	)
}
func tfConfigDataProjectByID(id string) string {
	return fmt.Sprintf(`
data "bitwarden_project" "foo_data" {
	provider	= bitwarden

	id 	= %s
}
`, id)
}

func tfConfigResourceProject(resourceName, projectName string) string {
	return fmt.Sprintf(`
	resource "bitwarden_project" "%s" {
		provider = bitwarden

		name = "%s"
	}
`, resourceName, projectName)
}
