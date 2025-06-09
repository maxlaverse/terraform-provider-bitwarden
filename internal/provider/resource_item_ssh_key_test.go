//go:build integration

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func TestAccResourceItemSSHKey(t *testing.T) {
	SkipIfVaultwardenBackend(t)

	ensureTestConfigurationReady(t)

	resourceName := "bitwarden_item_ssh_key.foo"
	var objectID string

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfigPasswordManagerProvider(testAccountFullAdmin) + tfConfigResourceItemSSHKey(),
				Check: resource.ComposeTestCheckFunc(
					checkItemSSHKey(resourceName),
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

const (
	examplePrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACDJmsEaDg9BOyXrpoEjCoFxJ1E49U8WBDuLEbTniBrf0AAAAIi9EgqMvRIK
jAAAAAtzc2gtZWQyNTUxOQAAACDJmsEaDg9BOyXrpoEjCoFxJ1E49U8WBDuLEbTniBrf0A
AAAECZuFCzMsNNgIveGNF+woHffSkplvs6ocy5b8ImW4MUWsmawRoOD0E7JeumgSMKgXEn
UTj1TxYEO4sRtOeIGt/QAAAAAAECAwQF
-----END OPENSSH PRIVATE KEY-----`
)

func tfConfigResourceItemSSHKey() string {
	return fmt.Sprintf(`
	resource "bitwarden_item_ssh_key" "foo" {
		provider 			= bitwarden

		organization_id     = "%s"
		collection_ids		= ["%s"]
		folder_id 			= "%s"
		name     			= "ssh-key-bar"
		reprompt			= true
		notes 				= "notes"
		private_key			= <<-EOT
%s
EOT
		public_key			= "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMmawRoOD0E7JeumgSMKgXEnUTj1TxYEO4sRtOeIGt/Q"
		key_fingerprint		= "SHA256:HevYQd7S5p2hGNApy4vYbPgCcbMiErDaSoRaEXVBz/w"

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
	}
`, testConfiguration.Resources.OrganizationID, testConfiguration.Resources.CollectionID, testConfiguration.Resources.FolderID, examplePrivateKey)
}

func checkItemSSHKey(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		checkItemBase(resourceName),
		resource.TestCheckResourceAttr(
			resourceName, schema_definition.AttributeSSHKeyPrivateKey, fmt.Sprintf("%s\n", examplePrivateKey),
		),
		resource.TestCheckResourceAttr(
			resourceName, schema_definition.AttributeSSHKeyPublicKey, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMmawRoOD0E7JeumgSMKgXEnUTj1TxYEO4sRtOeIGt/Q",
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeSSHKeyKeyFingerprint, regexp.MustCompile("^SHA256:HevYQd7S5p2hGNApy4vYbPgCcbMiErDaSoRaEXVBz/w$"),
		),
	)
}
