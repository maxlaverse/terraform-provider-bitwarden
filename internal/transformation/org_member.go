package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func OrganizationMemberObjectToSchema(ctx context.Context, obj *models.OrgMember, d *schema.ResourceData) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgMember)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeEmail, obj.Email)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeOrganizationID, obj.OrganizationId)
	if err != nil {
		return err
	}

	return nil
}

func OrganizationMemberToObject(ctx context.Context, d *schema.ResourceData) models.OrgMember {
	var obj models.OrgMember

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeEmail).(string); ok {
		obj.Email = v
	}

	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		obj.OrganizationId = v
	}

	return obj
}
