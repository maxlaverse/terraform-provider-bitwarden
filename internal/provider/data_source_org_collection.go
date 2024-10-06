package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func dataSourceOrgCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing organization collection.",
		ReadContext: resourceReadDataSourceObject(models.ObjectTypeOrgCollection),
		Schema:      orgCollectionSchema(DataSource),
	}
}
