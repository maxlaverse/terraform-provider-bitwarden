package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type applyOperationFn[T any] func(ctx context.Context, secret T) (*T, error)
type deleteOperationFn[T any] func(context.Context, T) error
type listOperationFn[T any] func(ctx context.Context, objType models.ObjectType, options ...bitwarden.ListObjectsOption) ([]T, error)

// TransformationOperation
type schemaToObjectTransformation[T any] func(ctx context.Context, d *schema.ResourceData) T
type objectToSchemaTransformation[T any] func(ctx context.Context, d *schema.ResourceData, obj *T) error

func applyOperation[T any](ctx context.Context, d *schema.ResourceData, clientOperation applyOperationFn[T], fromSchemaToObj schemaToObjectTransformation[T], fromObjToSchema objectToSchemaTransformation[T]) error {
	obj, err := clientOperation(ctx, fromSchemaToObj(ctx, d))
	if err != nil {
		return err
	}

	return fromObjToSchema(ctx, d, obj)
}

func searchOperation[T any](ctx context.Context, d *schema.ResourceData, clientOperation listOperationFn[T], fromObjToSchema objectToSchemaTransformation[T]) error {
	objType, ok := d.GetOk(schema_definition.AttributeObject)
	if !ok {
		return fmt.Errorf("BUG: object type not set in the resource data")
	}

	objs, err := clientOperation(ctx, models.ObjectType(objType.(string)), transformation.ListOptionsFromData(d)...)
	if err != nil {
		return err
	}

	if len(objs) == 0 {
		return fmt.Errorf("no object found matching the filter")
	} else if len(objs) > 1 {
		tflog.Warn(ctx, "Too many objects found", map[string]interface{}{"objects": objs})

		return fmt.Errorf("too many objects found")
	}

	obj := objs[0]

	// If the object exists but is marked as soft deleted, we return an error. This shouldn't happen
	// in theory since we never pass the --trash flag to the Bitwarden CLI when listing objects.
	switch object := any(obj).(type) {
	case *models.Object:
		if object.DeletedDate != nil {
			return errors.New("object is soft deleted")
		}
	}

	return fromObjToSchema(ctx, d, &obj)
}

func withNilReturn[T any](operation deleteOperationFn[T]) func(ctx context.Context, secret T) (*T, error) {
	return func(ctx context.Context, secret T) (*T, error) {
		return nil, operation(ctx, secret)
	}
}

func ignoreMissing(ctx context.Context, d *schema.ResourceData, err error) diag.Diagnostics {
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

func resourceImporter(stateContext schema.StateContextFunc) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: stateContext,
	}
}
