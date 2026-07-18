package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func OrganizationGroupObjectToSchema(ctx context.Context, obj *models.OrgGroup, d *schema.ResourceData) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(schema_definition.AttributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeOrganizationID, obj.OrganizationID)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeMember, obj.MemberIDs)
	if err != nil {
		return err
	}

	return nil
}

func OrganizationGroupToObject(ctx context.Context, d *schema.ResourceData) models.OrgGroup {
	var obj models.OrgGroup

	obj.ID = d.Id()

	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		obj.OrganizationID = v
	}

	if v, ok := d.Get(schema_definition.AttributeMember).(*schema.Set); ok {
		obj.MemberIDs = make([]string, v.Len())
		for k, v2 := range v.List() {
			obj.MemberIDs[k] = v2.(string)
		}
	}

	return obj
}
