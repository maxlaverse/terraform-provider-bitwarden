package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func resourceItemAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "Use this resource to create a file attachment within an item in Bitwarden.",

		CreateContext: resourceItemAttachmentCreate,
		ReadContext:   objectRead,
		DeleteContext: objectDelete,
		Importer:      importItemAttachmentResource(),

		Schema: map[string]*schema.Schema{
			attributeID: {
				Description: descriptionIdentifier,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeItemAttachmentFile: {
				Description: descriptionItemAttachmentFile,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			attributeItemAttachmentFileName: {
				Description: descriptionItemAttachmentFile,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeItemAttachmentItemID: {
				Description: descriptionIdentifier,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			attributeObject: {
				Description: descriptionInternal,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceItemAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(attributeObject, bw.ObjectTypeItemAttachment)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.Get(attributeItemAttachmentFile).(string); ok {
		err = d.Set(attributeItemAttachmentFileName, filepath.Base(v))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return objectCreate(ctx, d, meta)
}

func importItemAttachmentResource() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			d.SetId(d.Id())

			att, err := meta.(bw.Client).GetItemAttachment(d.Id())
			if err != nil {
				return nil, fmt.Errorf("could not find attachment with id %s: %v", d.Id(), err)
			}

			err = d.Set(attributeObject, bw.ObjectTypeItemAttachment)
			if err != nil {
				return nil, err
			}
			err = d.Set(attributeItemAttachmentItemID, att.ItemId)
			if err != nil {
				return nil, err
			}
			err = d.Set(attributeItemAttachmentFile, att.FileName)
			if err != nil {
				return nil, err
			}
			err = d.Set(attributeItemAttachmentFileName, att.FileName)
			if err != nil {
				return nil, err
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}
