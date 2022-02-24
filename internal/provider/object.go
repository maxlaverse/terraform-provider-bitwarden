package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

func objectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, meta.(bitwarden.Client).CreateObject)
}

func objectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, meta.(bitwarden.Client).GetObject)
}

func objectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, meta.(bitwarden.Client).EditObject)
}

func objectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, func(secret bitwarden.Object) (*bitwarden.Object, error) {
		return nil, meta.(bitwarden.Client).RemoveObject(secret)
	})
}

func objectOperation(ctx context.Context, d *schema.ResourceData, operation func(secret bitwarden.Object) (*bitwarden.Object, error)) diag.Diagnostics {
	obj, err := operation(objectStructFromData(d))
	if err != nil {
		if errors.Is(err, bitwarden.ErrNotFound) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(objectDataFromStruct(d, obj))
}

func objectDataFromStruct(d *schema.ResourceData, obj *bitwarden.Object) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(attributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(attributeObject, obj.Object)
	if err != nil {
		return err
	}

	if obj.Object == bitwarden.ObjectTypeItem {
		err = d.Set(attributeFolderID, obj.FolderID)
		if err != nil {
			return err
		}

		err = d.Set(attributeType, obj.Type)
		if err != nil {
			return err
		}

		err = d.Set(attributeNotes, obj.Notes)
		if err != nil {
			return err
		}

		err = d.Set(attributeField, objectFieldDataFromStruct(obj))
		if err != nil {
			return err
		}

		err = d.Set(attributeReprompt, obj.Reprompt == 1)
		if err != nil {
			return err
		}

		err = d.Set(attributeRevisionDate, obj.RevisionDate.Format(bitwarden.RevisionDateLayout))
		if err != nil {
			return err
		}

		if obj.Type == bitwarden.ItemTypeLogin {
			err = d.Set(attributeLoginPassword, obj.Login.Password)
			if err != nil {
				return err
			}

			err = d.Set(attributeLoginTotp, obj.Login.Totp)
			if err != nil {
				return err
			}

			err = d.Set(attributeLoginUsername, obj.Login.Username)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func objectStructFromData(d *schema.ResourceData) bitwarden.Object {
	var obj bitwarden.Object

	obj.ID = d.Id()
	if v, ok := d.Get(attributeName).(string); ok {
		obj.Name = v
	}

	if v, ok := d.Get(attributeObject).(string); ok {
		obj.Object = bitwarden.ObjectType(v)
	}

	if obj.Object == bitwarden.ObjectTypeItem {
		if v, ok := d.Get(attributeType).(int); ok {
			obj.Type = bitwarden.ItemType(v)
		}

		if v, ok := d.Get(attributeFolderID).(string); ok {
			obj.FolderID = v
		}

		if v, ok := d.Get(attributeFavorite).(bool); ok {
			obj.Favorite = v
		}

		if v, ok := d.Get(attributeNotes).(string); ok {
			obj.Notes = v
		}

		if v, ok := d.Get(attributeReprompt).(bool); ok && v {
			obj.Reprompt = 1
		}

		if v, ok := d.Get(attributeField).([]interface{}); ok {
			obj.Fields = objectFieldStructFromData(v)
		}

		if obj.Type == bitwarden.ItemTypeLogin {
			if v, ok := d.Get(attributeLoginPassword).(string); ok {
				obj.Login.Password = v
			}
			if v, ok := d.Get(attributeLoginTotp).(string); ok {
				obj.Login.Totp = v
			}
			if v, ok := d.Get(attributeLoginUsername).(string); ok {
				obj.Login.Username = v
			}
		}
	}

	return obj
}

func objectFieldDataFromStruct(obj *bitwarden.Object) []interface{} {
	fields := make([]interface{}, len(obj.Fields))
	for k, f := range obj.Fields {
		field := map[string]interface{}{
			attributeFieldName: f.Name,
		}
		if f.Type == bitwarden.FieldTypeText {
			field[attributeFieldText] = f.Value
		} else if f.Type == bitwarden.FieldTypeBoolean {
			field[attributeFieldBoolean] = (f.Value == "true")
		} else if f.Type == bitwarden.FieldTypeHidden {
			field[attributeFieldHidden] = f.Value
		} else if f.Type == bitwarden.FieldTypeLinked {
			field[attributeFieldLinked] = f.Value
		}
		fields[k] = field
	}
	return fields
}

func objectFieldStructFromData(vList []interface{}) []bitwarden.Field {
	fields := make([]bitwarden.Field, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		fields[k] = bitwarden.Field{
			Name: vc[attributeFieldName].(string),
		}
		if vs, ok := vc[attributeFieldText].(string); ok && len(vs) > 0 {
			fields[k].Type = bitwarden.FieldTypeText
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldHidden].(string); ok && len(vs) > 0 {
			fields[k].Type = bitwarden.FieldTypeHidden
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldLinked].(string); ok && len(vs) > 0 {
			fields[k].Type = bitwarden.FieldTypeLinked
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldBoolean].(bool); ok {
			fields[k].Type = bitwarden.FieldTypeBoolean
			if vs {
				fields[k].Value = "true"
			} else {
				fields[k].Value = "false"
			}
		}
	}
	return fields
}
