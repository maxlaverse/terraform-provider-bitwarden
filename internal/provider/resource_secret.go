package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecret() *schema.Resource {
	resourceSecretSchema := secretSchema(Resource)

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
