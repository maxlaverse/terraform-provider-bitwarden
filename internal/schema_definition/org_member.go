package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func OrgMemberSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeID: {
			Description: DescriptionIdentifier,
			Type:        schema.TypeString,
			Computed:    schemaType == Resource,
			Optional:    schemaType == DataSource,
		},
		AttributeOrganizationID: {
			Description: DescriptionOrganizationID,
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    schemaType == Resource,
		},
		AttributeName: {
			Description: DescriptionName,
			Type:        schema.TypeString,
			Computed:    true,
		},
		AttributeRole: {
			Description: DescriptionOrgMemberRole,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			ForceNew:    schemaType == Resource,
		},
	}

	if schemaType == Resource {
		base[AttributeEmail] = &schema.Schema{
			Description: DescriptionEmail,
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		}
		base[AttributeRole].Default = OrgMemberRoleUser
		base[AttributeRole].ValidateDiagFunc = validation.ToDiagFunc(validation.StringInSlice(
			[]string{OrgMemberRoleOwner, OrgMemberRoleAdmin, OrgMemberRoleUser, OrgMemberRoleManager}, false))
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
