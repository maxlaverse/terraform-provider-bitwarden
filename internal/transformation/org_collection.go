package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func OrganizationCollectionObjectToSchema(ctx context.Context, obj *models.OrgCollection, d *schema.ResourceData) error {
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

	users := make([]interface{}, len(obj.Users))
	for k, v := range obj.Users {
		users[k] = map[string]interface{}{
			schema_definition.AttributeCollectionMemberHidePasswords: v.HidePasswords,
			schema_definition.AttributeCollectionMemberOrgMemberId:   v.OrgMemberId,
			schema_definition.AttributeCollectionMemberReadOnly:      v.ReadOnly,
			schema_definition.AttributeCollectionMemberUserEmail:     v.UserEmail,
		}
	}

	err = d.Set(schema_definition.AttributeMember, users)
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

	if v, ok := d.Get(schema_definition.AttributeMember).([]interface{}); ok {
		obj.Users = make([]models.OrgCollectionMember, len(v))
		for k, v2 := range v {
			obj.Users[k] = models.OrgCollectionMember{
				HidePasswords: v2.(map[string]interface{})[schema_definition.AttributeCollectionMemberHidePasswords].(bool),
				ReadOnly:      v2.(map[string]interface{})[schema_definition.AttributeCollectionMemberReadOnly].(bool),
				UserEmail:     v2.(map[string]interface{})[schema_definition.AttributeCollectionMemberUserEmail].(string),

				// Note: We don't set OrgMemberId on purpose as it's computed and we're always going to lookup by email.
			}
		}
	}
	return obj
}
