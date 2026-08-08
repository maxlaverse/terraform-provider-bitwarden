package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemClientOp func(context.Context, models.Item) (*models.Item, error)

func itemApply(ctx context.Context, itemType models.ItemType, attr *transformation.MapData, op itemClientOp) error {
	obj, err := op(ctx, transformation.ItemSchemaToObject(itemType)(ctx, attr))
	if err != nil {
		return err
	}
	return transformation.ItemObjectToSchema(ctx, obj, attr)
}

func itemGet(ctx context.Context, bwClient bitwarden.PasswordManager, itemType models.ItemType, attr *transformation.MapData) error {
	obj, err := bwClient.GetItem(ctx, transformation.ItemSchemaToObject(itemType)(ctx, attr))
	if err != nil {
		return err
	}
	return transformation.ItemObjectToSchema(ctx, obj, attr)
}

func itemSearch(ctx context.Context, bwClient bitwarden.PasswordManager, itemType models.ItemType, attr *transformation.MapData) error {
	filters := append(transformation.ListOptionsFromData(attr), bitwarden.WithItemType(int(itemType)))
	obj, err := bwClient.FindItem(ctx, filters...)
	if err != nil {
		return err
	}
	return transformation.ItemObjectToSchema(ctx, obj, attr)
}

func itemIsSoftDeleted(attr *transformation.MapData) bool {
	v, ok := attr.Values()[schema_definition.AttributeDeletedDate].(string)
	return ok && len(v) > 0
}

type itemResource[T any] struct {
	clients        *ProviderClients
	typeNameSuffix string
	itemType       models.ItemType
	schema         func() rsschema.Schema
	toData         func(context.Context, T) *transformation.MapData
	fromData       func(*transformation.MapData) T
}

func (r *itemResource[T]) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.typeNameSuffix
}

func (r *itemResource[T]) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema()
}

func (r *itemResource[T]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	r.clients = clients
}

func (r *itemResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.toData(ctx, plan)
	if err := itemApply(ctx, r.itemType, attr, bwClient.CreateItem); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, r.fromData(attr))...)
}

func (r *itemResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.toData(ctx, state)
	if err := itemGet(ctx, bwClient, r.itemType, attr); err != nil {
		if errors.Is(err, models.ErrObjectNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		addErr(&resp.Diagnostics, err)
		return
	}

	if itemIsSoftDeleted(attr) {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, r.fromData(attr))...)
}

func (r *itemResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.toData(ctx, plan)
	if err := itemApply(ctx, r.itemType, attr, bwClient.EditItem); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, r.fromData(attr))...)
}

func (r *itemResource[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr := r.toData(ctx, state)
	if err := bwClient.DeleteItem(ctx, transformation.ItemSchemaToObject(r.itemType)(ctx, attr)); err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}
}

func (r *itemResource[T]) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(schema_definition.AttributeID), req, resp)
}

type itemDataSource[T any] struct {
	clients        *ProviderClients
	typeNameSuffix string
	itemType       models.ItemType
	schema         func() dsschema.Schema
	// prepareAttr builds the lookup MapData (including filters) from config and
	// returns the optional explicit object ID for Get vs Search.
	prepareAttr func(context.Context, T) (attr *transformation.MapData, objectID string)
	// resultToState merges the API-backed MapData with the original config
	// (typically to preserve filter attributes in state).
	resultToState func(attr *transformation.MapData, cfg T) T
}

func (d *itemDataSource[T]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + d.typeNameSuffix
}

func (d *itemDataSource[T]) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schema()
}

func (d *itemDataSource[T]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}
	d.clients = clients
}

func (d *itemDataSource[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg T
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bwClient, ok := requirePasswordManager(d.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	attr, objectID := d.prepareAttr(ctx, cfg)

	var err error
	if objectID != "" {
		err = itemGet(ctx, bwClient, d.itemType, attr)
	} else {
		err = itemSearch(ctx, bwClient, d.itemType, attr)
	}
	if err != nil {
		addErr(&resp.Diagnostics, err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, d.resultToState(attr, cfg))...)
}
