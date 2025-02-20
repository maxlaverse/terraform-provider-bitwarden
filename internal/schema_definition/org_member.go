package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func OrgMemberSchema() map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Optional:    true,
		},
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
		},
		AttributeEmail: {
			Description:  DescriptionEmail,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeEmail, AttributeID},
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return base
}
