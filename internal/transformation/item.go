package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func ItemObjectToSchema(ctx context.Context, obj *models.Item, d *schema.ResourceData) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(schema_definition.AttributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeObject, models.ObjectTypeItem)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeFolderID, obj.FolderID)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeType, obj.Type)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeNotes, obj.Notes)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeOrganizationID, obj.OrganizationID)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeFavorite, obj.Favorite)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeCollectionIDs, obj.CollectionIds)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeAttachments, ItemAttachmentsFromStruct(obj.Attachments))
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeField, ItemFieldDataFromStruct(obj))
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeReprompt, obj.Reprompt == 1)
	if err != nil {
		return err
	}

	if obj.RevisionDate != nil {
		err = d.Set(schema_definition.AttributeRevisionDate, obj.RevisionDate.Format(models.DateLayout))
		if err != nil {
			return err
		}
	}

	if obj.CreationDate != nil {
		err = d.Set(schema_definition.AttributeCreationDate, obj.CreationDate.Format(models.DateLayout))
		if err != nil {
			return err
		}
	}

	if obj.DeletedDate != nil {
		err = d.Set(schema_definition.AttributeDeletedDate, obj.DeletedDate.Format(models.DateLayout))
		if err != nil {
			return err
		}
	}

	if obj.Type == models.ItemTypeLogin {
		err = d.Set(schema_definition.AttributeLoginPassword, obj.Login.Password)
		if err != nil {
			return err
		}

		err = d.Set(schema_definition.AttributeLoginTotp, obj.Login.Totp)
		if err != nil {
			return err
		}

		err = d.Set(schema_definition.AttributeLoginUsername, obj.Login.Username)
		if err != nil {
			return err
		}

		err = d.Set(schema_definition.AttributeLoginURIs, objectLoginURIsFromStruct(ctx, obj.Login.URIs))
		if err != nil {
			return err
		}
	}
	return nil
}

func ItemSchemaToObject(ctx context.Context, d *schema.ResourceData) models.Item {
	var obj models.Item

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	obj.Object = models.ObjectTypeItem

	if v, ok := d.Get(schema_definition.AttributeType).(int); ok {
		obj.Type = models.ItemType(v)
	}

	if v, ok := d.Get(schema_definition.AttributeFolderID).(string); ok {
		obj.FolderID = v
	}

	if v, ok := d.Get(schema_definition.AttributeFavorite).(bool); ok && v {
		obj.Favorite = true
	}

	if v, ok := d.Get(schema_definition.AttributeNotes).(string); ok {
		obj.Notes = v
	}

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		obj.OrganizationID = v
	}

	if v, ok := d.Get(schema_definition.AttributeReprompt).(bool); ok && v {
		obj.Reprompt = 1
	}

	if vList, ok := d.Get(schema_definition.AttributeCollectionIDs).([]interface{}); ok {
		obj.CollectionIds = make([]string, len(vList))
		for k, v := range vList {
			obj.CollectionIds[k] = v.(string)
		}
	}

	if vList, ok := d.Get(schema_definition.AttributeAttachments).([]interface{}); ok {
		obj.Attachments = ItemAttachmentStructFromData(vList)
	}

	if v, ok := d.Get(schema_definition.AttributeField).([]interface{}); ok {
		obj.Fields = ObjectFieldStructFromData(v)
	}

	if obj.Type == models.ItemTypeLogin {
		if v, ok := d.Get(schema_definition.AttributeLoginPassword).(string); ok {
			obj.Login.Password = v
		}
		if v, ok := d.Get(schema_definition.AttributeLoginTotp).(string); ok {
			obj.Login.Totp = v
		}
		if v, ok := d.Get(schema_definition.AttributeLoginUsername).(string); ok {
			obj.Login.Username = v
		}
		if vList, ok := d.Get(schema_definition.AttributeLoginURIs).([]interface{}); ok {
			obj.Login.URIs = ObjectLoginURIsFromData(ctx, vList)
		}
	}

	return obj
}

func ItemFieldDataFromStruct(obj *models.Item) []interface{} {
	fields := make([]interface{}, len(obj.Fields))
	for k, f := range obj.Fields {
		field := map[string]interface{}{
			schema_definition.AttributeFieldName: f.Name,
		}
		if f.Type == models.FieldTypeText {
			field[schema_definition.AttributeFieldText] = f.Value
		} else if f.Type == models.FieldTypeBoolean {
			field[schema_definition.AttributeFieldBoolean] = (f.Value == "true")
		} else if f.Type == models.FieldTypeHidden {
			field[schema_definition.AttributeFieldHidden] = f.Value
		} else if f.Type == models.FieldTypeLinked {
			field[schema_definition.AttributeFieldLinked] = f.Value
		}
		fields[k] = field
	}
	return fields
}

func ItemAttachmentStructFromData(vList []interface{}) []models.Attachment {
	attachments := make([]models.Attachment, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		attachments[k] = models.Attachment{
			ID:       vc[schema_definition.AttributeID].(string),
			FileName: vc[schema_definition.AttributeAttachmentFileName].(string),
			Size:     vc[schema_definition.AttributeAttachmentSize].(string),
			SizeName: vc[schema_definition.AttributeAttachmentSizeName].(string),
			Url:      vc[schema_definition.AttributeAttachmentURL].(string),
		}
	}
	return attachments
}

func ItemAttachmentsFromStruct(objAttachments []models.Attachment) []interface{} {
	attachments := make([]interface{}, len(objAttachments))
	for k, f := range objAttachments {
		attachments[k] = map[string]interface{}{
			schema_definition.AttributeID:                 f.ID,
			schema_definition.AttributeAttachmentFileName: f.FileName,
			schema_definition.AttributeAttachmentSize:     f.Size,
			schema_definition.AttributeAttachmentSizeName: f.SizeName,
			schema_definition.AttributeAttachmentURL:      f.Url,
		}
	}
	return attachments
}

func ObjectFieldStructFromData(vList []interface{}) []models.Field {
	fields := make([]models.Field, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		fields[k] = models.Field{
			Name: vc[schema_definition.AttributeFieldName].(string),
		}
		if vs, ok := vc[schema_definition.AttributeFieldText].(string); ok && len(vs) > 0 {
			fields[k].Type = models.FieldTypeText
			fields[k].Value = vs
		} else if vs, ok := vc[schema_definition.AttributeFieldHidden].(string); ok && len(vs) > 0 {
			fields[k].Type = models.FieldTypeHidden
			fields[k].Value = vs
		} else if vs, ok := vc[schema_definition.AttributeFieldLinked].(string); ok && len(vs) > 0 {
			fields[k].Type = models.FieldTypeLinked
			fields[k].Value = vs
		} else if vs, ok := vc[schema_definition.AttributeFieldBoolean].(bool); ok {
			fields[k].Type = models.FieldTypeBoolean
			if vs {
				fields[k].Value = "true"
			} else {
				fields[k].Value = "false"
			}
		}
	}
	return fields
}

func ObjectLoginURIsFromData(ctx context.Context, vList []interface{}) []models.LoginURI {
	uris := make([]models.LoginURI, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		uris[k] = models.LoginURI{
			Match: schema_definition.StrMatchToInt(ctx, vc[schema_definition.AttributeLoginURIsMatch].(string)),
			URI:   vc[schema_definition.AttributeLoginURIsValue].(string),
		}
	}
	return uris
}

func objectLoginURIsFromStruct(ctx context.Context, objUris []models.LoginURI) []interface{} {
	uris := make([]interface{}, len(objUris))
	for k, f := range objUris {
		uris[k] = map[string]interface{}{
			schema_definition.AttributeLoginURIsMatch: schema_definition.IntMatchToStr(ctx, f.Match),
			schema_definition.AttributeLoginURIsValue: f.URI,
		}
	}
	return uris
}
