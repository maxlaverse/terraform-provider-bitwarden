package transformation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func AttachmentDataFromStruct(d *schema.ResourceData, attachment models.Attachment) error {
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
