package transformation

import (
	"context"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func OrganizationCollectionObjectToSchema(ctx context.Context, obj *models.OrgCollection, d AttrData) error {
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

	// Pass a plain slice; *schema.ResourceData.Set wraps TypeSet attributes.
	err = d.Set(schema_definition.AttributeMember, users)
	if err != nil {
		return err
	}

	groups := make([]interface{}, len(obj.Groups))
	for k, v := range obj.Groups {
		groups[k] = map[string]interface{}{
			schema_definition.AttributeCollectionMemberHidePasswords: v.HidePasswords,
			schema_definition.AttributeID:                            v.Id,
			schema_definition.AttributeCollectionMemberReadOnly:      v.ReadOnly,
			schema_definition.AttributeCollectionMemberManage:        v.Manage,
		}
	}

	err = d.Set(schema_definition.AttributeMemberGroup, groups)
	if err != nil {
		return err
	}

	return nil
}

func OrganizationCollectionToObject(ctx context.Context, d AttrData) models.OrgCollection {
	var obj models.OrgCollection

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	obj.Object = models.ObjectTypeOrgCollection

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		obj.OrganizationID = v
	}

	obj.Users = orgCollectionMembersFromData(d.Get(schema_definition.AttributeMember))
	obj.Groups = orgCollectionMembersFromData(d.Get(schema_definition.AttributeMemberGroup))

	return obj
}

func orgCollectionMembersFromData(v interface{}) []models.OrgCollectionMember {
	vList, ok := asInterfaceList(v)
	if !ok {
		return []models.OrgCollectionMember{}
	}

	members := make([]models.OrgCollectionMember, len(vList))
	for k, v2 := range vList {
		m, ok := v2.(map[string]interface{})
		if !ok {
			continue
		}
		members[k] = models.OrgCollectionMember{
			HidePasswords: boolFromMap(m, schema_definition.AttributeCollectionMemberHidePasswords),
			Id:            stringFromMap(m, schema_definition.AttributeID),
			ReadOnly:      boolFromMap(m, schema_definition.AttributeCollectionMemberReadOnly),
			Manage:        boolFromMap(m, schema_definition.AttributeCollectionMemberManage),
		}
	}
	return members
}

func stringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func boolFromMap(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
