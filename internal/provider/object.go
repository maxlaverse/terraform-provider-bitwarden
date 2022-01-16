package provider

import (
	"context"

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

		if v, ok := d.Get(attributeNotes).(string); ok {
			obj.Notes = v
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
