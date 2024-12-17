package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func dataSourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing folder.",
		ReadContext: withPasswordManager(resourceReadDataSourceObject(models.ObjectTypeFolder)),
		Schema:      schema_definition.FolderSchema(schema_definition.DataSource),
	}
}
