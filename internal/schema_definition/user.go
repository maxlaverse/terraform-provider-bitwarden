package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func UserSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    schemaType == DataSource,
		},
	}

	if schemaType == Resource {
		base[AttributeEmail] = &schema.Schema{
			Description: DescriptionEmail,
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		}
	} else {
		base[AttributeEmail] = &schema.Schema{
			Description:  DescriptionEmail,
			Type:         schema.TypeString,
			Optional:     true,
			AtLeastOneOf: []string{AttributeEmail, AttributeID},
		}
	}

	return base
}
