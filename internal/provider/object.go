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
)

func objectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, meta.(bitwarden.PasswordManager).CreateObject))
}

func objectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if _, idProvided := d.GetOk(attributeID); !idProvided {
		return diag.FromErr(objectSearch(ctx, d, meta))
	}

	return diag.FromErr(objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		obj, err := meta.(bitwarden.PasswordManager).GetObject(ctx, secret)
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

func objectSearch(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	objType, ok := d.GetOk(attributeObject)
	if !ok {
		return fmt.Errorf("BUG: object type not set in the resource data")
	}

	objs, err := meta.(bitwarden.PasswordManager).ListObjects(ctx, models.ObjectType(objType.(string)), listOptionsFromData(d)...)
	if err != nil {
		return err
	}

	// If the object is an item, also filter by type to avoid returning a login when a secure note is expected.
	if models.ObjectType(objType.(string)) == models.ObjectTypeItem {
		itemType, ok := d.GetOk(attributeType)
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
		attributeFilterSearch:         bitwarden.WithSearch,
		attributeFilterCollectionId:   bitwarden.WithCollectionID,
		attributeOrganizationID:       bitwarden.WithOrganizationID,
		attributeFilterFolderID:       bitwarden.WithFolderID,
		attributeFilterOrganizationID: bitwarden.WithOrganizationID,
		attributeFilterURL:            bitwarden.WithUrl,
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

func objectReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		return meta.(bitwarden.PasswordManager).GetObject(ctx, secret)
	})

	if errors.Is(err, models.ErrObjectNotFound) {
		d.SetId("")
		tflog.Warn(ctx, "Object not found, removing from state")
		return diag.Diagnostics{}
	}

	if _, exists := d.GetOk(attributeDeletedDate); exists {
		d.SetId("")
		tflog.Warn(ctx, "Object was soft deleted, removing from state")
		return diag.Diagnostics{}
	}

	return diag.FromErr(err)
}

func objectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, meta.(bitwarden.PasswordManager).EditObject))
}

func objectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(objectOperation(ctx, d, func(ctx context.Context, secret models.Object) (*models.Object, error) {
		return nil, meta.(bitwarden.PasswordManager).DeleteObject(ctx, secret)
	}))
}

func objectOperation(ctx context.Context, d *schema.ResourceData, operation func(ctx context.Context, secret models.Object) (*models.Object, error)) error {
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

	err := d.Set(attributeName, obj.Name)
	if err != nil {
		return err
	}

	err = d.Set(attributeObject, obj.Object)
	if err != nil {
		return err
	}

	// Object-specific fields
	switch obj.Object {
	case models.ObjectTypeOrgCollection:
		err = d.Set(attributeOrganizationID, obj.OrganizationID)
		if err != nil {
			return err
		}

	case models.ObjectTypeItem:
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

		err = d.Set(attributeOrganizationID, obj.OrganizationID)
		if err != nil {
			return err
		}

		err = d.Set(attributeFavorite, obj.Favorite)
		if err != nil {
			return err
		}

		err = d.Set(attributeCollectionIDs, obj.CollectionIds)
		if err != nil {
			return err
		}

		err = d.Set(attributeAttachments, objectAttachmentsFromStruct(obj.Attachments))
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

		if obj.RevisionDate != nil {
			err = d.Set(attributeRevisionDate, obj.RevisionDate.Format(models.DateLayout))
			if err != nil {
				return err
			}
		}

		if obj.CreationDate != nil {
			err = d.Set(attributeCreationDate, obj.CreationDate.Format(models.DateLayout))
			if err != nil {
				return err
			}
		}

		if obj.DeletedDate != nil {
			err = d.Set(attributeDeletedDate, obj.DeletedDate.Format(models.DateLayout))
			if err != nil {
				return err
			}
		}

		if obj.Type == models.ItemTypeLogin {
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

			err = d.Set(attributeLoginURIs, objectLoginURIsFromStruct(ctx, obj.Login.URIs))
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
	if v, ok := d.Get(attributeName).(string); ok {
		obj.Name = v
	}

	if v, ok := d.Get(attributeObject).(string); ok {
		obj.Object = models.ObjectType(v)
	}

	// Object-specific fields
	switch obj.Object {
	case models.ObjectTypeOrgCollection:
		if v, ok := d.Get(attributeOrganizationID).(string); ok {
			obj.OrganizationID = v
		}

	case models.ObjectTypeItem:
		if v, ok := d.Get(attributeType).(int); ok {
			obj.Type = models.ItemType(v)
		}

		if v, ok := d.Get(attributeFolderID).(string); ok {
			obj.FolderID = v
		}

		if v, ok := d.Get(attributeFavorite).(bool); ok && v {
			obj.Favorite = true
		}

		if v, ok := d.Get(attributeNotes).(string); ok {
			obj.Notes = v
		}

		if v, ok := d.Get(attributeOrganizationID).(string); ok {
			obj.OrganizationID = v
		}

		if v, ok := d.Get(attributeReprompt).(bool); ok && v {
			obj.Reprompt = 1
		}

		if vList, ok := d.Get(attributeCollectionIDs).([]interface{}); ok {
			obj.CollectionIds = make([]string, len(vList))
			for k, v := range vList {
				obj.CollectionIds[k] = v.(string)
			}
		}

		if vList, ok := d.Get(attributeAttachments).([]interface{}); ok {
			obj.Attachments = objectAttachmentStructFromData(vList)
		}

		if v, ok := d.Get(attributeField).([]interface{}); ok {
			obj.Fields = objectFieldStructFromData(v)
		}

		if obj.Type == models.ItemTypeLogin {
			if v, ok := d.Get(attributeLoginPassword).(string); ok {
				obj.Login.Password = v
			}
			if v, ok := d.Get(attributeLoginTotp).(string); ok {
				obj.Login.Totp = v
			}
			if v, ok := d.Get(attributeLoginUsername).(string); ok {
				obj.Login.Username = v
			}
			if vList, ok := d.Get(attributeLoginURIs).([]interface{}); ok {
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
			attributeFieldName: f.Name,
		}
		if f.Type == models.FieldTypeText {
			field[attributeFieldText] = f.Value
		} else if f.Type == models.FieldTypeBoolean {
			field[attributeFieldBoolean] = (f.Value == "true")
		} else if f.Type == models.FieldTypeHidden {
			field[attributeFieldHidden] = f.Value
		} else if f.Type == models.FieldTypeLinked {
			field[attributeFieldLinked] = f.Value
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
			ID:       vc[attributeID].(string),
			FileName: vc[attributeAttachmentFileName].(string),
			Size:     vc[attributeAttachmentSize].(string),
			SizeName: vc[attributeAttachmentSizeName].(string),
			Url:      vc[attributeAttachmentURL].(string),
		}
	}
	return attachments
}

func objectAttachmentsFromStruct(objAttachments []models.Attachment) []interface{} {
	attachments := make([]interface{}, len(objAttachments))
	for k, f := range objAttachments {
		attachments[k] = map[string]interface{}{
			attributeID:                 f.ID,
			attributeAttachmentFileName: f.FileName,
			attributeAttachmentSize:     f.Size,
			attributeAttachmentSizeName: f.SizeName,
			attributeAttachmentURL:      f.Url,
		}
	}
	return attachments
}

func objectFieldStructFromData(vList []interface{}) []models.Field {
	fields := make([]models.Field, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		fields[k] = models.Field{
			Name: vc[attributeFieldName].(string),
		}
		if vs, ok := vc[attributeFieldText].(string); ok && len(vs) > 0 {
			fields[k].Type = models.FieldTypeText
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldHidden].(string); ok && len(vs) > 0 {
			fields[k].Type = models.FieldTypeHidden
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldLinked].(string); ok && len(vs) > 0 {
			fields[k].Type = models.FieldTypeLinked
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldBoolean].(bool); ok {
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
			Match: strMatchToInt(ctx, vc[attributeLoginURIsMatch].(string)),
			URI:   vc[attributeLoginURIsValue].(string),
		}
	}
	return uris
}

func objectLoginURIsFromStruct(ctx context.Context, objUris []models.LoginURI) []interface{} {
	uris := make([]interface{}, len(objUris))
	for k, f := range objUris {
		uris[k] = map[string]interface{}{
			attributeLoginURIsMatch: intMatchToStr(ctx, f.Match),
			attributeLoginURIsValue: f.URI,
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
