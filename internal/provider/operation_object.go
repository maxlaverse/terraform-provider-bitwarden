package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

type objectOperationFunc func(ctx context.Context, secret models.Object) (*models.Object, error)

func opObjectCreate(attrObject models.ObjectType, attrType models.ItemType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		err := d.Set(schema_definition.AttributeObject, attrObject)
		if err != nil {
			return diag.FromErr(err)
		}
		err = d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return diag.FromErr(err)
		}

		return objectCreate(ctx, d, bwClient)
	}
}

func opObjectDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		return nil, bwClient.DeleteObject(ctx, secret)
	}))
}

func opObjectImport(attrObject models.ObjectType, attrType models.ItemType) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		d.SetId(d.Id())
		err := d.Set(schema_definition.AttributeObject, attrObject)
		if err != nil {
			return nil, err
		}
		err = d.Set(schema_definition.AttributeType, attrType)
		if err != nil {
			return nil, err
		}
		return []*schema.ResourceData{d}, nil
	}
}

func opObjectRead(objType models.ObjectType) passwordManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
		d.SetId(d.Get(schema_definition.AttributeID).(string))
		err := d.Set(schema_definition.AttributeObject, objType)
		if err != nil {
			return diag.FromErr(err)
		}
		return objectRead(ctx, d, bwClient)
	}
}

func opObjectReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	err := objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		return bwClient.GetObject(ctx, secret)
	})

	if errors.Is(err, models.ErrObjectNotFound) {
		d.SetId("")
		tflog.Warn(ctx, "Object not found, removing from state")
		return diag.Diagnostics{}
	}

	if _, exists := d.GetOk(schema_definition.AttributeDeletedDate); exists {
		d.SetId("")
		tflog.Warn(ctx, "Object was soft deleted, removing from state")
		return diag.Diagnostics{}
	}

	return diag.FromErr(err)
}

func opObjectUpdate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, bwClient.EditObject))
}

func objectCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, bwClient.CreateObject))
}

func objectRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	if _, idProvided := d.GetOk(schema_definition.AttributeID); !idProvided {
		return diag.FromErr(objectSearch(ctx, d, bwClient))
	}

	return diag.FromErr(objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		obj, err := bwClient.GetObject(ctx, secret)
		if obj != nil {
			// If the object exists but is marked as soft deleted, we return an error, because relying
			// on an object in the 'trash' sounds like a bad idea.
			if obj.DeletedDate != nil {
				return nil, errors.New("object is soft deleted")
			}

			if obj.ID != secret.ID {
				return nil, errors.New("returned object ID does not match requested object ID")
			}

			if obj.Type != secret.Type {
				return nil, errors.New("returned object type does not match requested object type")
			}
		}

		return obj, err
	}))
}

func objectSearch(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) error {
	objType, ok := d.GetOk(schema_definition.AttributeObject)
	if !ok {
		return fmt.Errorf("BUG: object type not set in the resource data")
	}

	objs, err := bwClient.ListObjects(ctx, models.ObjectType(objType.(string)), listOptionsFromData(d)...)
	if err != nil {
		return err
	}

	// If the object is an item, also filter by type to avoid returning a login when a secure note is expected.
	if models.ObjectType(objType.(string)) == models.ObjectTypeItem {
		itemType, ok := d.GetOk(schema_definition.AttributeType)
		if !ok {
			return fmt.Errorf("BUG: item type not set in the resource data")
		}

		objs = bwcli.FilterObjectsByType(objs, models.ItemType(itemType.(int)))
	}

	if len(objs) == 0 {
		return fmt.Errorf("no object found matching the filter")
	} else if len(objs) > 1 {
		objects := []string{}
		for _, obj := range objs {
			objects = append(objects, fmt.Sprintf("%s (%s)", obj.Name, obj.ID))
		}
		tflog.Warn(ctx, "Too many objects found", map[string]interface{}{"objects": objects})

		return fmt.Errorf("too many objects found")
	}

	obj := objs[0]

	// If the object exists but is marked as soft deleted, we return an error. This shouldn't happen
	// in theory since we never pass the --trash flag to the Bitwarden CLI when listing objects.
	if obj.DeletedDate != nil {
		return errors.New("object is soft deleted")
	}

	return objectDataFromStruct(ctx, d, &obj)
}

func objectOperation(ctx context.Context, d *schema.ResourceData, operation objectOperationFunc) error {
	obj, err := operation(ctx, objectStructFromData(ctx, d))
	if err != nil {
		return err
	}

	return objectDataFromStruct(ctx, d, obj)
}
