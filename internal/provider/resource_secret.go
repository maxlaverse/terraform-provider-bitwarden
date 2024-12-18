package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a secret.",
		CreateContext: withSecretsManager(opSecretCreate),
		ReadContext:   withSecretsManager(opSecretReadIgnoreMissing),
		UpdateContext: withSecretsManager(opSecretUpdate),
		DeleteContext: withSecretsManager(opSecretDelete),
		Schema:        schema_definition.SecretSchema(schema_definition.Resource),
		Importer:      resourceImporter(opSecretImport),
	}
}
