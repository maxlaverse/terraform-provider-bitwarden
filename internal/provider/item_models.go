package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

type itemCommonModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CollectionIDs  types.Set    `tfsdk:"collection_ids"`
	FolderID       types.String `tfsdk:"folder_id"`
	Notes          types.String `tfsdk:"notes"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Reprompt       types.Bool   `tfsdk:"reprompt"`
	CreationDate   types.String `tfsdk:"creation_date"`
	DeletedDate    types.String `tfsdk:"deleted_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
	Field          types.List   `tfsdk:"field"`
}

// itemVaultModel extends the common item with favorite/attachments (login and
// secure note; SSH keys omit these attributes in schema).
type itemVaultModel struct {
	itemCommonModel
	Favorite    types.Bool `tfsdk:"favorite"`
	Attachments types.List `tfsdk:"attachments"`
}

type itemFilterModel struct {
	FilterCollectionID   types.String `tfsdk:"filter_collection_id"`
	FilterFolderID       types.String `tfsdk:"filter_folder_id"`
	FilterOrganizationID types.String `tfsdk:"filter_organization_id"`
	Search               types.String `tfsdk:"search"`
}

func (m itemCommonModel) toDataMap(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		schema_definition.AttributeName:           m.Name.ValueString(),
		schema_definition.AttributeFolderID:       m.FolderID.ValueString(),
		schema_definition.AttributeNotes:          m.Notes.ValueString(),
		schema_definition.AttributeOrganizationID: m.OrganizationID.ValueString(),
		schema_definition.AttributeReprompt:       m.Reprompt.ValueBool(),
		schema_definition.AttributeCollectionIDs:  collectionIDsToStrings(ctx, m.CollectionIDs),
		schema_definition.AttributeField:          fieldListToData(m.Field),
	}
}

func itemCommonFromValues(id string, values map[string]interface{}) itemCommonModel {
	return itemCommonModel{
		ID:             types.StringValue(id),
		Name:           mapStr(values[schema_definition.AttributeName]),
		CollectionIDs:  stringsToSet(values[schema_definition.AttributeCollectionIDs]),
		FolderID:       mapStr(values[schema_definition.AttributeFolderID]),
		Notes:          mapStr(values[schema_definition.AttributeNotes]),
		OrganizationID: mapStr(values[schema_definition.AttributeOrganizationID]),
		Reprompt:       mapBool(values[schema_definition.AttributeReprompt]),
		CreationDate:   mapStr(values[schema_definition.AttributeCreationDate]),
		DeletedDate:    mapStr(values[schema_definition.AttributeDeletedDate]),
		RevisionDate:   mapStr(values[schema_definition.AttributeRevisionDate]),
		Field:          fieldDataToList(values[schema_definition.AttributeField]),
	}
}

func (m itemVaultModel) toDataMap(ctx context.Context) map[string]interface{} {
	out := m.itemCommonModel.toDataMap(ctx)
	out[schema_definition.AttributeFavorite] = m.Favorite.ValueBool()
	out[schema_definition.AttributeAttachments] = attachmentsListToData(m.Attachments)
	return out
}

func itemVaultFromValues(id string, values map[string]interface{}) itemVaultModel {
	return itemVaultModel{
		itemCommonModel: itemCommonFromValues(id, values),
		Favorite:        mapBool(values[schema_definition.AttributeFavorite]),
		Attachments:     attachmentsDataToList(values[schema_definition.AttributeAttachments]),
	}
}

func (f itemFilterModel) applyTo(attr *transformation.MapData) {
	_ = attr.Set(schema_definition.AttributeFilterCollectionId, f.FilterCollectionID.ValueString())
	_ = attr.Set(schema_definition.AttributeFilterFolderID, f.FilterFolderID.ValueString())
	_ = attr.Set(schema_definition.AttributeFilterOrganizationID, f.FilterOrganizationID.ValueString())
	_ = attr.Set(schema_definition.AttributeFilterSearch, f.Search.ValueString())
}

func newItemMapData(id string, values map[string]interface{}) *transformation.MapData {
	attr := transformation.NewMapData(values)
	attr.SetId(id)
	return attr
}
