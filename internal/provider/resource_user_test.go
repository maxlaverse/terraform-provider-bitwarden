//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccResourceUser(t *testing.T) {
	SkipIfOfficialBackend(t, "bitwarden_user uses Vaultwarden's admin API, which has no bitwarden.com equivalent")
	SkipIfOfficialCLI(t, "bitwarden_user is only supported by the embedded client")

	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_user.foo_user"
	testEmail := fmt.Sprintf("user-acc-%s@laverse.net", testConfiguration.UniqueTestIdentifier)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				ResourceName: resourceName,
				Config:       tfConfigPasswordManagerProviderWithAdminToken(testAccountFullAdmin) + tfConfigResourceUser(testEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, schema_definition.AttributeEmail, testEmail),
					resource.TestMatchResourceAttr(resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccResourceUser_ErrorsWithoutAdminToken(t *testing.T) {
	SkipIfOfficialBackend(t, "bitwarden_user uses Vaultwarden's admin API, which has no bitwarden.com equivalent")
	SkipIfOfficialCLI(t, "bitwarden_user is only supported by the embedded client")

	ensureTestConfigurationReady(t)

	testEmail := fmt.Sprintf("user-noadmin-%s@laverse.net", testConfiguration.UniqueTestIdentifier)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceUser(testEmail),
				ExpectError: regexp.MustCompile("admin_token must be configured"),
			},
		},
	})
}

func tfConfigResourceUser(email string) string {
	return fmt.Sprintf(`
resource "bitwarden_user" "foo_user" {
	email = "%s"
}
`, email)
}
