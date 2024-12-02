package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func orgCollectionSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		attributeID: {
			Description: descriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		attributeName: {
			Description: descriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		attributeObject: {
			Description: descriptionInternal,
			Type:        schema.TypeString,
			Computed:    true,
		},
		attributeOrganizationID: {
			Description: descriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
		},
		attributeMember: {
			Description: "TODO: Fix description",
			Type:        schema.TypeList,
			Elem:        membershipElem(),
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   false,
		},
	}

	if schemaType == DataSource {
		base[attributeFilterSearch] = &schema.Schema{
			Description:  descriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{attributeFilterSearch, attributeID},
		}
	}

	return base
}

func membershipElem() *schema.Resource {
	// validMatchStr := []string{"default", "base_domain", "host", "start_with", "exact", "regexp", "never"}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"org_user_id": {
				Description: "descriptionLoginUriMatch",
				Type:        schema.TypeString,
				// Default:          validMatchStr[0],
				// ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validMatchStr, false)),
				Required: true,
			},
			"read_only": {
				Description: "descriptionLoginUriValue",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}
