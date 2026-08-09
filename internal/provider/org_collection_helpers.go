package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var membershipAttrTypes = map[string]attr.Type{
	schema_definition.AttributeID:                            types.StringType,
	schema_definition.AttributeCollectionMemberReadOnly:      types.BoolType,
	schema_definition.AttributeCollectionMemberHidePasswords: types.BoolType,
	schema_definition.AttributeCollectionMemberManage:        types.BoolType,
}

type membershipModel struct {
	ID            types.String `tfsdk:"id"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	HidePasswords types.Bool   `tfsdk:"hide_passwords"`
	Manage        types.Bool   `tfsdk:"manage"`
}

func membershipSetToData(s types.Set) []interface{} {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	out := make([]interface{}, 0, len(s.Elements()))
	for _, e := range s.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		out = append(out, map[string]interface{}{
			schema_definition.AttributeID:                            strFromAttr(attrs[schema_definition.AttributeID]),
			schema_definition.AttributeCollectionMemberReadOnly:      boolFromAttr(attrs[schema_definition.AttributeCollectionMemberReadOnly]),
			schema_definition.AttributeCollectionMemberHidePasswords: boolFromAttr(attrs[schema_definition.AttributeCollectionMemberHidePasswords]),
			schema_definition.AttributeCollectionMemberManage:        boolFromAttr(attrs[schema_definition.AttributeCollectionMemberManage]),
		})
	}
	return out
}

func membershipDataToSet(v interface{}) types.Set {
	data, _ := v.([]interface{})
	elems := make([]attr.Value, 0, len(data))
	for _, item := range data {
		m, _ := item.(map[string]interface{})
		elems = append(elems, types.ObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			schema_definition.AttributeID:                            mapStr(m[schema_definition.AttributeID]),
			schema_definition.AttributeCollectionMemberReadOnly:      mapBoolOrFalse(m[schema_definition.AttributeCollectionMemberReadOnly]),
			schema_definition.AttributeCollectionMemberHidePasswords: mapBoolOrFalse(m[schema_definition.AttributeCollectionMemberHidePasswords]),
			schema_definition.AttributeCollectionMemberManage:        mapBoolOrFalse(m[schema_definition.AttributeCollectionMemberManage]),
		}))
	}
	return types.SetValueMust(types.ObjectType{AttrTypes: membershipAttrTypes}, elems)
}

func mapBoolOrFalse(v interface{}) types.Bool {
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolValue(false)
}

// normalizeMembershipSet rebuilds a membership set from configuration, turning
// null/unknown permission flags into false. Used from ModifyPlan to work around
// terraform-plugin-framework #867 (Computed defaults inside SetNested* can
// drop non-default configured bools from the plan).
func normalizeMembershipSet(ctx context.Context, s types.Set) (types.Set, diag.Diagnostics) {
	objectType := types.ObjectType{AttrTypes: membershipAttrTypes}
	if s.IsNull() {
		return types.SetNull(objectType), nil
	}
	if s.IsUnknown() {
		return s, nil
	}

	var members []membershipModel
	diags := s.ElementsAs(ctx, &members, false)
	if diags.HasError() {
		return types.SetNull(objectType), diags
	}

	for i := range members {
		members[i].ReadOnly = boolOrFalse(members[i].ReadOnly)
		members[i].HidePasswords = boolOrFalse(members[i].HidePasswords)
		members[i].Manage = boolOrFalse(members[i].Manage)
	}

	out, d := types.SetValueFrom(ctx, objectType, members)
	diags.Append(d...)
	return out, diags
}

func boolOrFalse(v types.Bool) types.Bool {
	if v.IsNull() || v.IsUnknown() {
		return types.BoolValue(false)
	}
	return v
}
