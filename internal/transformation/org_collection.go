package transformation

import (
	"context"
	"hash/fnv"

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

	err = d.Set(schema_definition.AttributeOrganizationID, obj.OrganizationID)
	if err != nil {
		return err
	}

	users := make([]interface{}, len(obj.Users))
	for k, v := range obj.Users {
		users[k] = map[string]interface{}{
			schema_definition.AttributeCollectionMemberHidePasswords: v.HidePasswords,
			schema_definition.AttributeID:                            v.Id,
			schema_definition.AttributeCollectionMemberReadOnly:      v.ReadOnly,
			schema_definition.AttributeCollectionMemberManage:        v.Manage,
		}
	}

	set := schema.NewSet(func(i interface{}) int {
		return hashStringToInt(i.(map[string]interface{})[schema_definition.AttributeID].(string))
	}, users)
	err = d.Set(schema_definition.AttributeMember, set)
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

	if v, ok := d.Get(schema_definition.AttributeMember).(*schema.Set); ok {
		obj.Users = make([]models.OrgCollectionMember, v.Len())
		for k, v2 := range v.List() {
			obj.Users[k] = models.OrgCollectionMember{
				HidePasswords: v2.(map[string]interface{})[schema_definition.AttributeCollectionMemberHidePasswords].(bool),
				Id:            v2.(map[string]interface{})[schema_definition.AttributeID].(string),
				ReadOnly:      v2.(map[string]interface{})[schema_definition.AttributeCollectionMemberReadOnly].(bool),
				Manage:        v2.(map[string]interface{})[schema_definition.AttributeCollectionMemberManage].(bool),
			}
		}
	}

	return obj
}

func hashStringToInt(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
