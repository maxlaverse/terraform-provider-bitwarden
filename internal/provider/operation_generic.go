package provider

import (
	"context"
	"errors"

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
type findOperationFn[T any] func(ctx context.Context, options ...bitwarden.ListObjectsOption) (*T, error)

// TransformationOperation
type schemaToObjectTransformation[T any] func(ctx context.Context, d *schema.ResourceData) T
type objectToSchemaTransformation[T any] func(ctx context.Context, obj *T, d *schema.ResourceData) error

func applyOperation[T any](ctx context.Context, d *schema.ResourceData, clientOperation applyOperationFn[T], fromSchemaToObj schemaToObjectTransformation[T], fromObjToSchema objectToSchemaTransformation[T]) error {
	obj, err := clientOperation(ctx, fromSchemaToObj(ctx, d))
	if err != nil {
		return err
	}

	return fromObjToSchema(ctx, obj, d)
}

func searchOperation[T any](ctx context.Context, d *schema.ResourceData, clientOperation findOperationFn[T], fromObjToSchema objectToSchemaTransformation[T]) error {
	obj, err := clientOperation(ctx, transformation.ListOptionsFromData(d)...)
	if err != nil {
		return err
	}
	return fromObjToSchema(ctx, obj, d)
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
