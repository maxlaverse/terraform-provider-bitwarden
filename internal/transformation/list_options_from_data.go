package transformation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func ListOptionsFromData(d *schema.ResourceData) []bitwarden.ListObjectsOption {
	filters := []bitwarden.ListObjectsOption{}

	filterMap := map[string]bitwarden.ListObjectsOptionGenerator{
		schema_definition.AttributeFilterSearch:         bitwarden.WithSearch,
		schema_definition.AttributeFilterCollectionId:   bitwarden.WithCollectionID,
		schema_definition.AttributeOrganizationID:       bitwarden.WithOrganizationID,
		schema_definition.AttributeFilterFolderID:       bitwarden.WithFolderID,
		schema_definition.AttributeFilterOrganizationID: bitwarden.WithOrganizationID,
		schema_definition.AttributeFilterURL:            bitwarden.WithUrl,
	}

	for attribute, optionFunc := range filterMap {
		v, ok := d.GetOk(attribute)
		if !ok {
			continue
		}

		if v, ok := v.(string); ok && len(v) > 0 {
			filters = append(filters, optionFunc(v))
		}
	}

	return filters
}
