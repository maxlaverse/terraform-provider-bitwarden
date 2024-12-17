package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceProject() *schema.Resource {
	resourceProjectSchema := schema_definition.ProjectSchema(schema_definition.Resource)

	return &schema.Resource{
		Description:   "Manages a Project.",
		CreateContext: withSecretsManager(resourceCreateProject()),
		ReadContext:   withSecretsManager(resourceReadProjectIgnoreMissing),
		UpdateContext: withSecretsManager(resourceUpdateProject),
		DeleteContext: withSecretsManager(resourceDeleteProject),
		Schema:        resourceProjectSchema,
		Importer:      resourceImporter(resourceImportProject),
	}
}
