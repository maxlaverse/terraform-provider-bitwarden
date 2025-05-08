package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func OrgCollectionSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    true,
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Required:    schemaType == Resource,
		},
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
		},
		AttributeMember: {
			Description: DescriptionCollectionMember,
			Type:        schema.TypeSet,
			Elem:        membershipElem(),
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   false,
		},
	}

	if schemaType == DataSource {
		base[AttributeFilterSearch] = &schema.Schema{
			Description:  DescriptionFilterSearch,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeFilterSearch, AttributeID},
		}
	}

	return base
}

func membershipElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			AttributeID: {
				Description: DescriptionIdentifier,
				Type:        schema.TypeString,
				Required:    true,
			},
			AttributeCollectionMemberReadOnly: {
				Description: DescriptionCollectionMemberReadOnly,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			AttributeCollectionMemberHidePasswords: {
				Description: DescriptionCollectionMemberHidePasswords,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			AttributeCollectionMemberManage: {
				Description: DescriptionCollectionMemberManage,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}
