package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

// NewSDK returns the SDKv2 provider factory. During the Framework migration it is
// muxed with New() so existing resources/data sources keep working while
// Framework-native implementations are added incrementally.
//
// Environment-variable defaults and cross-field credential checks run in
// configure (applyProviderConfigEnvDefaults / validateProviderConfig), not via
// schema DefaultFunc / AtLeastOneOf, which are unreliable under mux and would
// reject env-only configuration.
func NewSDK(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				// Credential attributes (cross-field rules enforced in validateProviderConfig)
				schema_definition.AttributeMasterPassword: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionMasterPassword,
					Optional:    true,
					Sensitive:   true,
				},
				schema_definition.AttributeSessionKey: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionSessionKey,
					Optional:    true,
					Sensitive:   true,
				},
				schema_definition.AttributeClientID: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionClientID,
					Optional:    true,
				},
				schema_definition.AttributeClientSecret: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionClientSecret,
					Optional:    true,
					Sensitive:   true,
				},
				schema_definition.AttributeBwsAccessToken: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionBwsAccessToken,
					Optional:    true,
					Sensitive:   true,
				},

				// Standalone attributes
				schema_definition.AttributeServer: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionServer,
					Optional:    true,
				},
				schema_definition.AttributeProviderEmail: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionProviderEmail,
					Optional:    true,
				},
				schema_definition.AttributeVaultPath: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionVaultPath,
					Optional:    true,
				},
				schema_definition.AttributeExtraCACertsPath: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionExtraCACertsPath,
					Optional:    true,
				},
				schema_definition.AttributeClientImplementation: {
					Type:             schema.TypeString,
					Description:      schema_definition.DescriptionClientImplementation,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{schema_definition.ClientImplementationCLI, schema_definition.ClientImplementationEmbedded}, false)),
				},

				// Experimental
				schema_definition.AttributeExperimental: {
					Description: schema_definition.DescriptionExperimental,
					Type:        schema.TypeSet,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schema_definition.AttributeExperimentalEmbeddedClient: {
								Description: schema_definition.DescriptionExperimentalEmbeddedClient,
								Type:        schema.TypeBool,
								Optional:    true,
								Deprecated:  "Use client_implementation = \"embedded\" instead.",
								// Both may be set; experimental.embedded_client wins in getClientImplementation().
								// client_implementation is defaulted there when unset, so a schema-level conflict
								// cannot distinguish an explicit "cli" from the default.
							},
							schema_definition.AttributeExperimentalDisableSyncAfterWriteVerification: {
								Description: schema_definition.DescriptionExperimentalDisableSyncAfterWriteVerification,
								Type:        schema.TypeBool,
								Optional:    true,
							},
						},
					},
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"bitwarden_attachment":       dataSourceAttachment(),
				"bitwarden_item_login":       dataSourceItemLogin(),
				"bitwarden_item_secure_note": dataSourceItemSecureNote(),
				"bitwarden_item_ssh_key":     dataSourceItemSSHKey(),
				"bitwarden_org_collection":   dataSourceOrgCollection(),
				"bitwarden_secret":           dataSourceSecret(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"bitwarden_attachment":       resourceAttachment(),
				"bitwarden_item_login":       resourceItemLogin(),
				"bitwarden_item_secure_note": resourceItemSecureNote(),
				"bitwarden_item_ssh_key":     resourceItemSSHKey(),
				"bitwarden_org_collection":   resourceOrgCollection(),
				"bitwarden_secret":           resourceSecret(),
			},
		}

		p.ConfigureContextFunc = providerConfigureSDK(version)
		return p
	}
}
