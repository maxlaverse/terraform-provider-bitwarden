package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				// Attributes which depend on each other
				schema_definition.AttributeMasterPassword: {
					Type:          schema.TypeString,
					Description:   schema_definition.DescriptionMasterPassword,
					ConflictsWith: []string{schema_definition.AttributeSessionKey, schema_definition.AttributeBwsAccessToken},
					AtLeastOneOf:  []string{schema_definition.AttributeSessionKey, schema_definition.AttributeBwsAccessToken},
					Optional:      true,
					DefaultFunc:   schema.EnvDefaultFunc("BW_PASSWORD", nil),
				},
				schema_definition.AttributeSessionKey: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionSessionKey,
					AtLeastOneOf: []string{schema_definition.AttributeMasterPassword, schema_definition.AttributeBwsAccessToken},
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("BW_SESSION", nil),
				},
				schema_definition.AttributeClientID: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionClientID,
					Optional:     true,
					RequiredWith: []string{schema_definition.AttributeClientSecret, schema_definition.AttributeMasterPassword},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTID", nil),
				},
				schema_definition.AttributeClientSecret: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionClientSecret,
					Optional:     true,
					RequiredWith: []string{schema_definition.AttributeClientID, schema_definition.AttributeMasterPassword},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTSECRET", nil),
				},
				schema_definition.AttributeBwsAccessToken: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionBwsAccessToken,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BWS_ACCESS_TOKEN", nil),
				},

				// Standalone attributes
				schema_definition.AttributeServer: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionServer,
					Required:    true,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{"BW_URL", "BWS_SERVER_URL"}, bitwarden.DefaultBitwardenServerURL),
				},
				schema_definition.AttributeProviderEmail: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionProviderEmail,
					Optional:     true,
					AtLeastOneOf: []string{schema_definition.AttributeBwsAccessToken, schema_definition.AttributeClientID, schema_definition.AttributeSessionKey},
					DefaultFunc:  schema.EnvDefaultFunc("BW_EMAIL", nil),
				},
				schema_definition.AttributeVaultPath: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionVaultPath,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BITWARDENCLI_APPDATA_DIR", ".bitwarden/"),
				},
				schema_definition.AttributeExtraCACertsPath: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionExtraCACertsPath,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("NODE_EXTRA_CA_CERTS", nil),
				},
				schema_definition.AttributeClientImplementation: {
					Type:             schema.TypeString,
					Description:      schema_definition.DescriptionClientImplementation,
					Optional:         true,
					Default:          schema_definition.ClientImplementationCLI,
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
								// Note: We don't use ConflictsWith because client_implementation has a default value. To
								// properly detect if it was explicitly set (vs using the default) would require additional
								// code. Instead, we allow both to be set and let experimental.embedded_client take
								// precedence in getClientImplementation().
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
				"bitwarden_folder":           dataSourceFolder(),
				"bitwarden_item_login":       dataSourceItemLogin(),
				"bitwarden_item_secure_note": dataSourceItemSecureNote(),
				"bitwarden_item_ssh_key":     dataSourceItemSSHKey(),
				"bitwarden_org_collection":   dataSourceOrgCollection(),
				"bitwarden_org_group":        dataSourceOrgGroup(),
				"bitwarden_org_member":       dataSourceOrgMember(),
				"bitwarden_organization":     dataSourceOrganization(),
				"bitwarden_project":          dataSourceProject(),
				"bitwarden_secret":           dataSourceSecret(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"bitwarden_attachment":       resourceAttachment(),
				"bitwarden_folder":           resourceFolder(),
				"bitwarden_item_login":       resourceItemLogin(),
				"bitwarden_item_secure_note": resourceItemSecureNote(),
				"bitwarden_item_ssh_key":     resourceItemSSHKey(),
				"bitwarden_org_collection":   resourceOrgCollection(),
				"bitwarden_project":          resourceProject(),
				"bitwarden_secret":           resourceSecret(),
			},
		}

		p.ConfigureContextFunc = providerConfigure(version, p)
		return p
	}
}

