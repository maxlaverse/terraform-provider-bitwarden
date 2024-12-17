package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceSecret() *schema.Resource {
	resourceSecretSchema := schema_definition.SecretSchema(schema_definition.Resource)

	return &schema.Resource{
		Description:   "Manages a secret.",
		CreateContext: withSecretsManager(resourceCreateSecret()),
		ReadContext:   withSecretsManager(resourceReadSecretIgnoreMissing),
		UpdateContext: withSecretsManager(resourceUpdateSecret),
		DeleteContext: withSecretsManager(resourceDeleteSecret),
		Schema:        resourceSecretSchema,
		Importer:      resourceImporter(resourceImportSecret),
	}
}
