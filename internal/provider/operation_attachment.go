package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opAttachmentCreate(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	existingAttachments, err := listExistingAttachments(ctx, bwClient, itemId)
	if err != nil {
		return diag.FromErr(err)
	}

	var obj *models.Item
	filePath, fileSpecified := d.GetOk(schema_definition.AttributeAttachmentFile)
	content, contentSpecified := d.GetOk(schema_definition.AttributeAttachmentContent)
	fileName, fileNameSpecified := d.GetOk(schema_definition.AttributeAttachmentFileName)
	dynamicFilePath, dynamicFileSpecified := d.GetOk(schema_definition.AttributeAttachmentDynamicFile)
	_, hashSpecified := d.GetOk(schema_definition.AttributeAttachmentDataHash)

	if fileSpecified {
		obj, err = bwClient.CreateAttachmentFromFile(ctx, itemId, filePath.(string))
	} else if dynamicFileSpecified {
		if !hashSpecified {
			// if hash not explicitly set, populate it.
			hashErr := d.Set(schema_definition.AttributeAttachmentDataHash, fileHash(dynamicFilePath))
			if hashErr != nil {
				return diag.FromErr(hashErr)
			}
		}
		// Create the Attachment
		obj, err = bwClient.CreateAttachmentFromFile(ctx, itemId, dynamicFilePath.(string))
	} else if contentSpecified && fileNameSpecified {
		obj, err = bwClient.CreateAttachmentFromContent(ctx, itemId, fileName.(string), []byte(content.(string)))
	} else {
		err = errors.New("BUG: either file or content&file_name should be specified")
	}
	if err != nil {
		return diag.FromErr(err)
	}

	attachmentsRemoved, attachmentsAdded := compareLists(existingAttachments, obj.Attachments)
	if len(attachmentsAdded) == 0 {
		return diag.FromErr(errors.New("BUG: no attachment found after creation"))
	} else if len(attachmentsAdded) > 1 {
		return diag.FromErr(errors.New("BUG: more than one attachment created"))
	} else if len(attachmentsRemoved) > 1 {
		return diag.FromErr(errors.New("BUG: at least one attachment removed"))
	}

	return diag.FromErr(transformation.AttachmentObjectToSchema(ctx, attachmentsAdded[0], d))
}

func opAttachmentDelete(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	return diag.FromErr(bwClient.DeleteAttachment(ctx, itemId, d.Id()))
}

func opAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid ID specified, should be in the format <item_id>/<attachment_id>: '%s'", d.Id())
	}
	d.SetId(split[0])
	err := d.Set(schema_definition.AttributeAttachmentItemID, split[1])
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func opAttachmentRead(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	attachmentId := d.Get(schema_definition.AttributeID).(string)

	content, err := bwClient.GetAttachment(ctx, itemId, attachmentId)
	if err != nil {
		return diag.FromErr(err)
	}

	// Read is used for datasource, so set hash value automatically. Since datasource is content/file/dynamic upload
	// agnostic, always calculate the hash as it can be useful even if not using the dynamic upload feature.
	hashErr := d.Set(schema_definition.AttributeAttachmentDataHash, contentHash(string(content)))
	if hashErr != nil {
		return diag.FromErr(hashErr)
	}

	d.SetId(attachmentId)

	return diag.FromErr(d.Set(schema_definition.AttributeAttachmentContent, string(content)))
}

func opAttachmentReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)
	dynamicFilePath, dynamicFileSpecified := d.GetOk(schema_definition.AttributeAttachmentDynamicFile)
	_, hashSpecified := d.GetOk(schema_definition.AttributeAttachmentDataHash)
	if dynamicFileSpecified && !hashSpecified {
		// if dynamic_file is specified, but hash is not, populate it.
		hashErr := d.Set(schema_definition.AttributeAttachmentDataHash, fileHash(dynamicFilePath))
		if hashErr != nil {
			return diag.FromErr(hashErr)
		}
	}
	obj, err := bwClient.GetItem(ctx, models.Item{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		// If the item is not found, we can't simply consider the attachment as
		// deleted, because we won't have an item to attach it to.
		// This means we don't need a special handling for NotFound errors and
		// should just return whatever we get.
		return diag.FromErr(err)
	}

	for _, attachment := range obj.Attachments {
		if attachment.ID == d.Id() {
			return diag.FromErr(transformation.AttachmentObjectToSchema(ctx, attachment, d))
		}
	}

	// If the item exists but the attachment is not found, we consider the
	// attachment as deleted.
	d.SetId("")
	return diag.Diagnostics{}
}

func listExistingAttachments(ctx context.Context, client bitwarden.PasswordManager, itemId string) ([]models.Attachment, error) {
	obj, err := client.GetItem(ctx, models.Item{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		return nil, err
	}
	return obj.Attachments, nil
}

func compareLists(listA []models.Attachment, listB []models.Attachment) ([]models.Attachment, []models.Attachment) {
	return itemsOnlyInSecondList(listB, listA), itemsOnlyInSecondList(listA, listB)
}

func itemsOnlyInSecondList(firstList []models.Attachment, secondList []models.Attachment) []models.Attachment {
	result := []models.Attachment{}
	for _, secondAttachment := range secondList {
		found := false
		for _, firstAttachment := range firstList {
			if firstAttachment.ID == secondAttachment.ID {
				found = true
				break
			}
		}
		if !found {
			result = append(result, secondAttachment)
		}
	}
	return result
}

func fileSha1Sum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	outputChecksum := hash.Sum(nil)

	return hex.EncodeToString(outputChecksum[:]), nil
}

func contentSha1Sum(content string) (string, error) {
	hash := sha1.New()
	_, err := hash.Write([]byte(content))
	if err != nil {
		return "", err
	}
	outputChecksum := hash.Sum(nil)

	return hex.EncodeToString(outputChecksum[:]), nil
}

func contentHash(val interface{}) string {
	hash, _ := contentSha1Sum(val.(string))
	return hash
}

func fileHashComputable(val interface{}, _ cty.Path) diag.Diagnostics {
	_, err := fileSha1Sum(val.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to compute hash of file: %w", err))
	}
	return diag.Diagnostics{}
}

func fileHash(val interface{}) string {
	hash, _ := fileSha1Sum(val.(string))
	return hash
}
