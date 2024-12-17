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

func listOptionsFromData(d *schema.ResourceData) []bitwarden.ListObjectsOption {
	filters := []bitwarden.ListObjectsOption{}

	filterMap := map[string]bitwarden.ListObjectsOptionGenerator{
		schema_definition.AttributeFilterSearch:         bitwarden.WithSearch,
		schema_definition.AttributeFilterCollectionId:   bitwarden.WithCollectionID,
		schema_definition.AttributeOrganizationID:       bitwarden.WithOrganizationID,
		schema_definition.AttributeFilterFolderID:       bitwarden.WithFolderID,
		schema_definition.AttributeFilterOrganizationID: bitwarden.WithOrganizationID,
		schema_definition.AttributeFilterURL:            bitwarden.WithUrl,
	}

	for attribute, optionFunc := range filterMap {
		v, ok := d.GetOk(attribute)
		if !ok {
			continue
		}

		if v, ok := v.(string); ok && len(v) > 0 {
			filters = append(filters, optionFunc(v))
		}
	}
	return filters
}

func objectOperation(ctx context.Context, d *schema.ResourceData, operation objectOperationFunc) error {
	obj, err := operation(ctx, objectStructFromData(ctx, d))
	if err != nil {
		return err
	}

	return objectDataFromStruct(ctx, d, obj)
}

func objectDataFromStruct(ctx context.Context, d *schema.ResourceData, obj *models.Object) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(schema_definition.AttributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeObject, obj.Object)
	if err != nil {
		return err
	}

	// Object-specific fields
	switch obj.Object {
	case models.ObjectTypeOrgCollection:
		err = d.Set(schema_definition.AttributeOrganizationID, obj.OrganizationID)
		if err != nil {
			return err
		}

	case models.ObjectTypeItem:
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

		err = d.Set(schema_definition.AttributeAttachments, objectAttachmentsFromStruct(obj.Attachments))
		if err != nil {
			return err
		}

		err = d.Set(schema_definition.AttributeField, objectFieldDataFromStruct(obj))
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
	}

	return nil
}

func objectStructFromData(ctx context.Context, d *schema.ResourceData) models.Object {
	var obj models.Object

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	if v, ok := d.Get(schema_definition.AttributeObject).(string); ok {
		obj.Object = models.ObjectType(v)
	}

	// Object-specific fields
	switch obj.Object {
	case models.ObjectTypeOrgCollection:
		if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
			obj.OrganizationID = v
		}

	case models.ObjectTypeItem:
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
			obj.Attachments = objectAttachmentStructFromData(vList)
		}

		if v, ok := d.Get(schema_definition.AttributeField).([]interface{}); ok {
			obj.Fields = objectFieldStructFromData(v)
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
				obj.Login.URIs = objectLoginURIsFromData(ctx, vList)
			}
		}
	}

	return obj
}

func objectFieldDataFromStruct(obj *models.Object) []interface{} {
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

func objectAttachmentStructFromData(vList []interface{}) []models.Attachment {
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

func objectAttachmentsFromStruct(objAttachments []models.Attachment) []interface{} {
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

func objectFieldStructFromData(vList []interface{}) []models.Field {
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

func objectLoginURIsFromData(ctx context.Context, vList []interface{}) []models.LoginURI {
	uris := make([]models.LoginURI, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		uris[k] = models.LoginURI{
			Match: strMatchToInt(ctx, vc[schema_definition.AttributeLoginURIsMatch].(string)),
			URI:   vc[schema_definition.AttributeLoginURIsValue].(string),
		}
	}
	return uris
}

func objectLoginURIsFromStruct(ctx context.Context, objUris []models.LoginURI) []interface{} {
	uris := make([]interface{}, len(objUris))
	for k, f := range objUris {
		uris[k] = map[string]interface{}{
			schema_definition.AttributeLoginURIsMatch: intMatchToStr(ctx, f.Match),
			schema_definition.AttributeLoginURIsValue: f.URI,
		}
	}
	return uris
}

func intMatchToStr(ctx context.Context, match *models.URIMatch) URIMatchStr {
	if match == nil {
		return URIMatchDefault
	}

	switch *match {
	case models.URIMatchBaseDomain:
		return URIMatchBaseDomain
	case models.URIMatchHost:
		return URIMatchHost
	case models.URIMatchStartWith:
		return URIMatchStartWith
	case models.URIMatchExact:
		return URIMatchExact
	case models.URIMatchRegExp:
		return URIMatchRegExp
	case models.URIMatchNever:
		return URIMatchNever
	default:
		tflog.Warn(ctx, "unsupported integer value for URI match - Falling back to default", map[string]interface{}{"match": *match})
		return URIMatchDefault
	}
}

func strMatchToInt(ctx context.Context, match string) *models.URIMatch {
	var v models.URIMatch
	switch match {
	case string(URIMatchDefault):
		return nil
	case string(URIMatchBaseDomain):
		v = models.URIMatchBaseDomain
	case string(URIMatchHost):
		v = models.URIMatchHost
	case string(URIMatchStartWith):
		v = models.URIMatchStartWith
	case string(URIMatchExact):
		v = models.URIMatchExact
	case string(URIMatchRegExp):
		v = models.URIMatchRegExp
	case string(URIMatchNever):
		v = models.URIMatchNever
	default:
		tflog.Warn(ctx, "unsupported string value for URI match - Falling back to default", map[string]interface{}{"match": match})
		return nil
	}
	return &v
}

type URIMatchStr string

const (
	URIMatchDefault    URIMatchStr = "default"
	URIMatchBaseDomain URIMatchStr = "base_domain"
	URIMatchHost       URIMatchStr = "host"
	URIMatchStartWith  URIMatchStr = "start_with"
	URIMatchExact      URIMatchStr = "exact"
	URIMatchRegExp     URIMatchStr = "regexp"
	URIMatchNever      URIMatchStr = "never"
)
