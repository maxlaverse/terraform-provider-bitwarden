package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceCreateAttachment(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	existingAttachments, err := listExistingAttachments(ctx, bwClient, itemId)
	if err != nil {
		return diag.FromErr(err)
	}

	var obj *models.Object
	filePath, fileSpecified := d.GetOk(schema_definition.AttributeAttachmentFile)
	content, contentSpecified := d.GetOk(schema_definition.AttributeAttachmentContent)
	fileName, fileNameSpecified := d.GetOk(schema_definition.AttributeAttachmentFileName)
	if fileSpecified {
		obj, err = bwClient.CreateAttachmentFromFile(ctx, itemId, filePath.(string))
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

	return diag.FromErr(attachmentDataFromStruct(d, attachmentsAdded[0]))
}

func resourceReadAttachment(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	obj, err := bwClient.GetObject(ctx, models.Object{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		// If the item is not found, we can't simply consider the attachment as
		// deleted, because we won't have an item to attach it to.
		// This means we don't need a special handling for NotFound errors and
		// should just return whatever we get.
		return diag.FromErr(err)
	}

	for _, attachment := range obj.Attachments {
		if attachment.ID == d.Id() {
			return diag.FromErr(attachmentDataFromStruct(d, attachment))
		}
	}

	// If the item exists but the attachment is not found, we consider the
	// attachment as deleted.
	d.SetId("")
	return diag.Diagnostics{}
}

func resourceDeleteAttachment(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	return diag.FromErr(bwClient.DeleteAttachment(ctx, itemId, d.Id()))
}

func attachmentDataFromStruct(d *schema.ResourceData, attachment models.Attachment) error {
	d.SetId(attachment.ID)

	err := d.Set(schema_definition.AttributeAttachmentFileName, attachment.FileName)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeAttachmentSize, attachment.Size)
	if err != nil {
		return err
	}
	err = d.Set(schema_definition.AttributeAttachmentSizeName, attachment.SizeName)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeAttachmentURL, attachment.Url)
	if err != nil {
		return err
	}

	return nil
}

func resourceReadDataSourceAttachment(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) diag.Diagnostics {
	itemId := d.Get(schema_definition.AttributeAttachmentItemID).(string)

	attachmentId := d.Get(schema_definition.AttributeID).(string)

	content, err := bwClient.GetAttachment(ctx, itemId, attachmentId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(attachmentId)

	return diag.FromErr(d.Set(schema_definition.AttributeAttachmentContent, string(content)))
}

func listExistingAttachments(ctx context.Context, client bitwarden.PasswordManager, itemId string) ([]models.Attachment, error) {
	obj, err := client.GetObject(ctx, models.Object{ID: itemId, Object: models.ObjectTypeItem})
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
