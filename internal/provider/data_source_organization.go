package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information on an existing Organization.",
		ReadContext: readDataSourceObject(bw.ObjectTypeOrganization),
		Schema:      organizationSchema(),
	}
}
