package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func mapBool(v interface{}) types.Bool {
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

func strFromAttr(v attr.Value) string {
	if s, ok := v.(types.String); ok {
		return s.ValueString()
	}
	return ""
}

func boolFromAttr(v attr.Value) bool {
	if b, ok := v.(types.Bool); ok {
		return b.ValueBool()
	}
	return false
}

func toStringSlice(v interface{}) []string {
	switch vv := v.(type) {
	case []string:
		return vv
	case []interface{}:
		out := make([]string, 0, len(vv))
		for _, item := range vv {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func collectionIDsToStrings(ctx context.Context, s types.Set) []string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	var out []string
	s.ElementsAs(ctx, &out, false)
	return out
}

func stringsToSet(v interface{}) types.Set {
	strs := toStringSlice(v)
	elems := make([]attr.Value, 0, len(strs))
	for _, s := range strs {
		elems = append(elems, types.StringValue(s))
	}
	return types.SetValueMust(types.StringType, elems)
}

var fieldAttrTypes = map[string]attr.Type{
	schema_definition.AttributeFieldName:    types.StringType,
	schema_definition.AttributeFieldText:    types.StringType,
	schema_definition.AttributeFieldBoolean: types.BoolType,
	schema_definition.AttributeFieldHidden:  types.StringType,
	schema_definition.AttributeFieldLinked:  types.StringType,
}

func fieldListToData(l types.List) []interface{} {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]interface{}, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		out = append(out, map[string]interface{}{
			schema_definition.AttributeFieldName:    strFromAttr(attrs[schema_definition.AttributeFieldName]),
			schema_definition.AttributeFieldText:    strFromAttr(attrs[schema_definition.AttributeFieldText]),
			schema_definition.AttributeFieldBoolean: boolFromAttr(attrs[schema_definition.AttributeFieldBoolean]),
			schema_definition.AttributeFieldHidden:  strFromAttr(attrs[schema_definition.AttributeFieldHidden]),
			schema_definition.AttributeFieldLinked:  strFromAttr(attrs[schema_definition.AttributeFieldLinked]),
		})
	}
	return out
}

func fieldDataToList(v interface{}) types.List {
	data, _ := v.([]interface{})
	elems := make([]attr.Value, 0, len(data))
	for _, item := range data {
		m, _ := item.(map[string]interface{})
		obj := types.ObjectValueMust(fieldAttrTypes, map[string]attr.Value{
			schema_definition.AttributeFieldName:    mapStr(m[schema_definition.AttributeFieldName]),
			schema_definition.AttributeFieldText:    mapStr(m[schema_definition.AttributeFieldText]),
			schema_definition.AttributeFieldBoolean: mapBool(m[schema_definition.AttributeFieldBoolean]),
			schema_definition.AttributeFieldHidden:  mapStr(m[schema_definition.AttributeFieldHidden]),
			schema_definition.AttributeFieldLinked:  mapStr(m[schema_definition.AttributeFieldLinked]),
		})
		elems = append(elems, obj)
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: fieldAttrTypes}, elems)
}

var uriAttrTypes = map[string]attr.Type{
	schema_definition.AttributeLoginURIsMatch: types.StringType,
	schema_definition.AttributeLoginURIsValue: types.StringType,
}

func uriListToData(l types.List) []interface{} {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]interface{}, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		out = append(out, map[string]interface{}{
			schema_definition.AttributeLoginURIsMatch: strFromAttr(attrs[schema_definition.AttributeLoginURIsMatch]),
			schema_definition.AttributeLoginURIsValue: strFromAttr(attrs[schema_definition.AttributeLoginURIsValue]),
		})
	}
	return out
}

func uriDataToList(v interface{}) types.List {
	data, _ := v.([]interface{})
	elems := make([]attr.Value, 0, len(data))
	for _, item := range data {
		m, _ := item.(map[string]interface{})
		obj := types.ObjectValueMust(uriAttrTypes, map[string]attr.Value{
			schema_definition.AttributeLoginURIsMatch: mapStr(m[schema_definition.AttributeLoginURIsMatch]),
			schema_definition.AttributeLoginURIsValue: mapStr(m[schema_definition.AttributeLoginURIsValue]),
		})
		elems = append(elems, obj)
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: uriAttrTypes}, elems)
}

var attachmentAttrTypes = map[string]attr.Type{
	schema_definition.AttributeID:                 types.StringType,
	schema_definition.AttributeAttachmentFileName: types.StringType,
	schema_definition.AttributeAttachmentSize:     types.StringType,
	schema_definition.AttributeAttachmentSizeName: types.StringType,
	schema_definition.AttributeAttachmentURL:      types.StringType,
}

func attachmentsListToData(l types.List) []interface{} {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]interface{}, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		out = append(out, map[string]interface{}{
			schema_definition.AttributeID:                 strFromAttr(attrs[schema_definition.AttributeID]),
			schema_definition.AttributeAttachmentFileName: strFromAttr(attrs[schema_definition.AttributeAttachmentFileName]),
			schema_definition.AttributeAttachmentSize:     strFromAttr(attrs[schema_definition.AttributeAttachmentSize]),
			schema_definition.AttributeAttachmentSizeName: strFromAttr(attrs[schema_definition.AttributeAttachmentSizeName]),
			schema_definition.AttributeAttachmentURL:      strFromAttr(attrs[schema_definition.AttributeAttachmentURL]),
		})
	}
	return out
}

func attachmentsDataToList(v interface{}) types.List {
	data, _ := v.([]interface{})
	elems := make([]attr.Value, 0, len(data))
	for _, item := range data {
		m, _ := item.(map[string]interface{})
		obj := types.ObjectValueMust(attachmentAttrTypes, map[string]attr.Value{
			schema_definition.AttributeID:                 mapStr(m[schema_definition.AttributeID]),
			schema_definition.AttributeAttachmentFileName: mapStr(m[schema_definition.AttributeAttachmentFileName]),
			schema_definition.AttributeAttachmentSize:     mapStr(m[schema_definition.AttributeAttachmentSize]),
			schema_definition.AttributeAttachmentSizeName: mapStr(m[schema_definition.AttributeAttachmentSizeName]),
			schema_definition.AttributeAttachmentURL:      mapStr(m[schema_definition.AttributeAttachmentURL]),
		})
		elems = append(elems, obj)
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: attachmentAttrTypes}, elems)
}
