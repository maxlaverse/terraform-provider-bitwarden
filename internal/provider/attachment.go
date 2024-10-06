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
)

func attachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	itemId := d.Get(attributeAttachmentItemID).(string)

	existingAttachments, err := listExistingAttachments(ctx, meta.(bitwarden.PasswordManager), itemId)
	if err != nil {
		return diag.FromErr(err)
	}

	filePath := d.Get(attributeAttachmentFile).(string)
	obj, err := meta.(bitwarden.PasswordManager).CreateAttachment(ctx, itemId, filePath)
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

func attachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	itemId := d.Get(attributeAttachmentItemID).(string)

	obj, err := meta.(bitwarden.PasswordManager).GetObject(ctx, models.Object{ID: itemId, Object: models.ObjectTypeItem})
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

func attachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	itemId := d.Get(attributeAttachmentItemID).(string)

	return diag.FromErr(meta.(bitwarden.PasswordManager).DeleteAttachment(ctx, itemId, d.Id()))
}

func attachmentDataFromStruct(d *schema.ResourceData, attachment models.Attachment) error {
	d.SetId(attachment.ID)

	err := d.Set(attributeAttachmentFileName, attachment.FileName)
	if err != nil {
		return err
	}

	err = d.Set(attributeAttachmentSize, attachment.Size)
	if err != nil {
		return err
	}
	err = d.Set(attributeAttachmentSizeName, attachment.SizeName)
	if err != nil {
		return err
	}

	err = d.Set(attributeAttachmentURL, attachment.Url)
	if err != nil {
		return err
	}

	return nil
}

func readDataSourceAttachment() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		itemId := d.Get(attributeAttachmentItemID).(string)

		attachmentId := d.Get(attributeID).(string)

		content, err := meta.(bitwarden.PasswordManager).GetAttachment(ctx, itemId, attachmentId)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(attachmentId)

		return diag.FromErr(d.Set(attributeAttachmentContent, string(content)))
	}
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
