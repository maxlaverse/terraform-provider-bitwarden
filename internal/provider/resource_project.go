package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	resourceProjectSchema := projectSchema(Resource)

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
