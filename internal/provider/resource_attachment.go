package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAttachment() *schema.Resource {
	resourceAttachmentSchema := attachmentSchema()
	resourceAttachmentSchema[attributeAttachmentFile] = &schema.Schema{
		Description:      descriptionItemAttachmentFile,
		Type:             schema.TypeString,
		Optional:         true,
		ConflictsWith:    []string{attributeAttachmentContent},
		AtLeastOneOf:     []string{attributeAttachmentContent},
		ForceNew:         true,
		ValidateDiagFunc: fileHashComputable,
		StateFunc:        fileHash,
	}
	resourceAttachmentSchema[attributeAttachmentContent] = &schema.Schema{
		Description:   descriptionItemAttachmentFile,
		Type:          schema.TypeString,
		Optional:      true,
		RequiredWith:  []string{attributeAttachmentContent},
		ConflictsWith: []string{attributeAttachmentFile},
		AtLeastOneOf:  []string{attributeAttachmentFile},
		ForceNew:      true,
		StateFunc:     contentHash,
	}
	resourceAttachmentSchema[attributeAttachmentFileName] = &schema.Schema{
		Description:   descriptionItemAttachmentFileName,
		Type:          schema.TypeString,
		RequiredWith:  []string{attributeAttachmentContent},
		ConflictsWith: []string{attributeAttachmentFile},
		ComputedWhen:  []string{attributeAttachmentFile},
		AtLeastOneOf:  []string{attributeAttachmentFile},
		ForceNew:      true,
		Optional:      true,
		Computed:      true,
	}

	resourceAttachmentSchema[attributeAttachmentItemID] = &schema.Schema{
		Description: descriptionItemIdentifier,
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}

	return &schema.Resource{
		Description: "Manages an item attachment.",

		CreateContext: withPasswordManager(resourceCreateAttachment),
		ReadContext:   withPasswordManager(resourceReadAttachment),
		DeleteContext: withPasswordManager(resourceDeleteAttachment),
		Importer:      resourceImporter(resourceImportAttachment),

		Schema: resourceAttachmentSchema,
	}
}

func resourceImportAttachment(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid ID specified, should be in the format <item_id>/<attachment_id>: '%s'", d.Id())
	}
	d.SetId(split[0])
	d.Set(attributeAttachmentItemID, split[1])
	return []*schema.ResourceData{d}, nil
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
