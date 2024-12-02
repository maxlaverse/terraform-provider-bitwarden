package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func OrganizationCollectionObjectToSchema(ctx context.Context, d *schema.ResourceData, obj *models.OrgCollection) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(schema_definition.AttributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeObject, models.ObjectTypeOrgCollection)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeOrganizationID, obj.OrganizationID)
	if err != nil {
		return err
	}

	return nil
}

func OrganizationCollectionToObject(ctx context.Context, d *schema.ResourceData) models.OrgCollection {
	var obj models.OrgCollection

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	obj.Object = models.ObjectTypeOrgCollection

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		obj.OrganizationID = v
	}

	return obj
}
