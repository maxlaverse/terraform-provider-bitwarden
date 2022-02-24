package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceItemLogin(t *testing.T) {
	ensureVaultwardenConfigured(t)

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfTestProvider() + tfTestDataItemLogin(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.bitwarden_item_login.foo", attributeName, regexp.MustCompile("^login-([0-9]+)$"),
					),
					resource.TestMatchResourceAttr(
						"data.bitwarden_item_login.foo", attributeLoginUsername, regexp.MustCompile("^test-user$"),
					),
					resource.TestMatchResourceAttr(
						"data.bitwarden_item_login.foo", attributeLoginPassword, regexp.MustCompile("^test-password$"),
					),
				),
			},
		},
	})
}

func tfTestDataItemLogin() string {
	return fmt.Sprintf(`
data "bitwarden_item_login" "foo" {
  id = "%s"
}
`, testItemLoginID)
}
