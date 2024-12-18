package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Project.",
		CreateContext: withSecretsManager(opProjectCreate),
		ReadContext:   withSecretsManager(opProjectReadIgnoreMissing),
		UpdateContext: withSecretsManager(opProjectUpdate),
		DeleteContext: withSecretsManager(opProjectDelete),
		Schema:        schema_definition.ProjectSchema(schema_definition.Resource),
		Importer:      resourceImporter(opProjectImport),
	}
}
