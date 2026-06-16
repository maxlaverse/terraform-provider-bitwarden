//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccDataSourceUserAttribute(t *testing.T) {
	SkipIfOfficialBackend(t, "bitwarden_user uses Vaultwarden's admin API, which has no bitwarden.com equivalent")
	SkipIfOfficialCLI(t, "bitwarden_user is only supported by the embedded client")

	ensureTestConfigurationReady(t)

	testEmail := fmt.Sprintf("user-ds-%s@laverse.net", testConfiguration.UniqueTestIdentifier)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProviderWithAdminToken(testAccountFullAdmin) + fmt.Sprintf(`
resource "bitwarden_user" "ds_target" {
	email = "%s"
}

data "bitwarden_user" "by_email" {
	email = bitwarden_user.ds_target.email
}

data "bitwarden_user" "by_id" {
	id = bitwarden_user.ds_target.id
}
`, testEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.bitwarden_user.by_email", schema_definition.AttributeEmail, testEmail),
					resource.TestMatchResourceAttr("data.bitwarden_user.by_email", schema_definition.AttributeID, regexp.MustCompile(regExpId)),
					resource.TestCheckResourceAttrPair("data.bitwarden_user.by_email", schema_definition.AttributeID, "bitwarden_user.ds_target", schema_definition.AttributeID),
					resource.TestCheckResourceAttrPair("data.bitwarden_user.by_id", schema_definition.AttributeID, "bitwarden_user.ds_target", schema_definition.AttributeID),
					resource.TestCheckResourceAttr("data.bitwarden_user.by_id", schema_definition.AttributeEmail, testEmail),
				),
			},
		},
	})
}

func TestAccDataSourceUser_ErrorsWithoutAdminToken(t *testing.T) {
	SkipIfOfficialBackend(t, "bitwarden_user uses Vaultwarden's admin API, which has no bitwarden.com equivalent")
	SkipIfOfficialCLI(t, "bitwarden_user is only supported by the embedded client")

	ensureTestConfigurationReady(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + `
data "bitwarden_user" "missing_admin" {
	email = "anyone@example.com"
}
`,
				ExpectError: regexp.MustCompile("admin_token must be configured"),
			},
		},
	})
}

func TestAccDataSourceUser_NotFound(t *testing.T) {
	SkipIfOfficialBackend(t, "bitwarden_user uses Vaultwarden's admin API, which has no bitwarden.com equivalent")
	SkipIfOfficialCLI(t, "bitwarden_user is only supported by the embedded client")

	ensureTestConfigurationReady(t)

	missingEmail := fmt.Sprintf("does-not-exist-%s@laverse.net", testConfiguration.UniqueTestIdentifier)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProviderWithAdminToken(testAccountFullAdmin) + fmt.Sprintf(`
data "bitwarden_user" "missing" {
	email = "%s"
}
`, missingEmail),
				ExpectError: regexp.MustCompile("not found"),
			},
		},
	})
}
