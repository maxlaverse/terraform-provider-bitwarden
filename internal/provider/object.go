package provider

import (
	"context"
	"errors"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func objectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, meta.(bw.Client).CreateObject)
}

func objectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, meta.(bw.Client).GetObject)
}

func objectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, meta.(bw.Client).EditObject)
}

func objectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return objectOperation(ctx, d, func(secret bw.Object) (*bw.Object, error) {
		return nil, meta.(bw.Client).RemoveObject(secret)
	})
}

func objectOperation(ctx context.Context, d *schema.ResourceData, operation func(secret bw.Object) (*bw.Object, error)) diag.Diagnostics {
	obj, err := operation(objectStructFromData(d))

	if err != nil {
		if errors.Is(err, bw.ErrNotFound) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(objectDataFromStruct(d, obj))
}

func objectDataFromStruct(d *schema.ResourceData, obj *bw.Object) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(attributeObject, obj.Object)
	if err != nil {
		return err
	}

	if obj.Object == bw.ObjectTypeItemAttachment {
		err = d.Set(attributeItemAttachmentFile, obj.File)
		if err != nil {
			return err
		}

		err = d.Set(attributeItemAttachmentFileName, obj.FileName)
		if err != nil {
			return err
		}

		err = d.Set(attributeItemAttachmentItemID, obj.ItemId)
		if err != nil {
			return err
		}
	} else {
		err := d.Set(attributeName, obj.Name)
		if err != nil {
			return err
		}

		if obj.Object == bw.ObjectTypeItem {
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

			err = d.Set(attributeField, objectFieldDataFromStruct(obj))
			if err != nil {
				return err
			}

			err = d.Set(attributeReprompt, obj.Reprompt == 1)
			if err != nil {
				return err
			}

			err = d.Set(attributeRevisionDate, obj.RevisionDate.Format(bw.RevisionDateLayout))
			if err != nil {
				return err
			}

			if obj.Type == bw.ItemTypeLogin {
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

				err = d.Set(attributeLoginURIs, objectLoginURIsFromStruct(obj.Login.URIs))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func objectStructFromData(d *schema.ResourceData) bw.Object {
	var obj bw.Object

	obj.ID = d.Id()

	if v, ok := d.Get(attributeObject).(string); ok {
		obj.Object = bw.ObjectType(v)
	}

	if obj.Object == bw.ObjectTypeItemAttachment {
		if v, ok := d.Get(attributeItemAttachmentFile).(string); ok {
			obj.File = v
		}

		if v, ok := d.Get(attributeItemAttachmentFileName).(string); ok {
			obj.FileName = v
		}

		if v, ok := d.Get(attributeItemAttachmentItemID).(string); ok {
			obj.ItemId = v
		}
	} else {
		if v, ok := d.Get(attributeName).(string); ok {
			obj.Name = v
		}

		if obj.Object == bw.ObjectTypeItem {
			if v, ok := d.Get(attributeType).(int); ok {
				obj.Type = bw.ItemType(v)
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

			if v, ok := d.Get(attributeField).([]interface{}); ok {
				obj.Fields = objectFieldStructFromData(v)
			}

			if obj.Type == bw.ItemTypeLogin {
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
					obj.Login.URIs = objectLoginURIsFromData(vList)
				}
			}
		}
	}

	return obj
}

func objectFieldDataFromStruct(obj *bw.Object) []interface{} {
	fields := make([]interface{}, len(obj.Fields))
	for k, f := range obj.Fields {
		field := map[string]interface{}{
			attributeFieldName: f.Name,
		}
		if f.Type == bw.FieldTypeText {
			field[attributeFieldText] = f.Value
		} else if f.Type == bw.FieldTypeBoolean {
			field[attributeFieldBoolean] = (f.Value == "true")
		} else if f.Type == bw.FieldTypeHidden {
			field[attributeFieldHidden] = f.Value
		} else if f.Type == bw.FieldTypeLinked {
			field[attributeFieldLinked] = f.Value
		}
		fields[k] = field
	}
	return fields
}

func objectFieldStructFromData(vList []interface{}) []bw.Field {
	fields := make([]bw.Field, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		fields[k] = bw.Field{
			Name: vc[attributeFieldName].(string),
		}
		if vs, ok := vc[attributeFieldText].(string); ok && len(vs) > 0 {
			fields[k].Type = bw.FieldTypeText
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldHidden].(string); ok && len(vs) > 0 {
			fields[k].Type = bw.FieldTypeHidden
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldLinked].(string); ok && len(vs) > 0 {
			fields[k].Type = bw.FieldTypeLinked
			fields[k].Value = vs
		} else if vs, ok := vc[attributeFieldBoolean].(bool); ok {
			fields[k].Type = bw.FieldTypeBoolean
			if vs {
				fields[k].Value = "true"
			} else {
				fields[k].Value = "false"
			}
		}
	}
	return fields
}

func objectLoginURIsFromData(vList []interface{}) []bw.LoginURI {
	uris := make([]bw.LoginURI, len(vList))
	for k, v := range vList {
		vc := v.(map[string]interface{})
		uris[k] = bw.LoginURI{
			Match: strMatchToInt(vc[attributeLoginURIsMatch].(string)),
			URI:   vc[attributeLoginURIsValue].(string),
		}
	}
	return uris
}

func objectLoginURIsFromStruct(objUris []bw.LoginURI) []interface{} {
	uris := make([]interface{}, len(objUris))
	for k, f := range objUris {
		uris[k] = map[string]interface{}{
			attributeLoginURIsMatch: intMatchToStr(f.Match),
			attributeLoginURIsValue: f.URI,
		}
	}
	return uris
}

func intMatchToStr(match *bw.URIMatch) URIMatchStr {
	if match == nil {
		return URIMatchDefault
	}

	switch *match {
	case bw.URIMatchBaseDomain:
		return URIMatchBaseDomain
	case bw.URIMatchHost:
		return URIMatchHost
	case bw.URIMatchStartWith:
		return URIMatchStartWith
	case bw.URIMatchExact:
		return URIMatchExact
	case bw.URIMatchRegExp:
		return URIMatchRegExp
	case bw.URIMatchNever:
		return URIMatchNever
	default:
		log.Printf("unsupported integer value for URI match: '%d'. Falling back to default\n", *match)
		return URIMatchDefault
	}
}

func strMatchToInt(match string) *bw.URIMatch {
	var v bw.URIMatch
	switch match {
	case string(URIMatchDefault):
		return nil
	case string(URIMatchBaseDomain):
		v = bw.URIMatchBaseDomain
	case string(URIMatchHost):
		v = bw.URIMatchHost
	case string(URIMatchStartWith):
		v = bw.URIMatchStartWith
	case string(URIMatchExact):
		v = bw.URIMatchExact
	case string(URIMatchRegExp):
		v = bw.URIMatchRegExp
	case string(URIMatchNever):
		v = bw.URIMatchNever
	default:
		log.Printf("unsupported string value for URI match: '%s'. Falling back to default\n", match)
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
