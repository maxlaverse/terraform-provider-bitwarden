package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceAttachment() *schema.Resource {
	resourceAttachmentSchema := schema_definition.AttachmentSchema()
	resourceAttachmentSchema[schema_definition.AttributeAttachmentFile] = &schema.Schema{
		Description:      schema_definition.DescriptionItemAttachmentFile,
		Type:             schema.TypeString,
		Optional:         true,
		ConflictsWith:    []string{schema_definition.AttributeAttachmentContent},
		AtLeastOneOf:     []string{schema_definition.AttributeAttachmentContent},
		ForceNew:         true,
		ValidateDiagFunc: fileHashComputable,
		StateFunc:        fileHash,
	}
	resourceAttachmentSchema[schema_definition.AttributeAttachmentContent] = &schema.Schema{
		Description:   schema_definition.DescriptionItemAttachmentFile,
		Type:          schema.TypeString,
		Optional:      true,
		RequiredWith:  []string{schema_definition.AttributeAttachmentContent},
		ConflictsWith: []string{schema_definition.AttributeAttachmentFile},
		AtLeastOneOf:  []string{schema_definition.AttributeAttachmentFile},
		ForceNew:      true,
		StateFunc:     contentHash,
	}
	resourceAttachmentSchema[schema_definition.AttributeAttachmentFileName] = &schema.Schema{
		Description:   schema_definition.DescriptionItemAttachmentFileName,
		Type:          schema.TypeString,
		RequiredWith:  []string{schema_definition.AttributeAttachmentContent},
		ConflictsWith: []string{schema_definition.AttributeAttachmentFile},
		ComputedWhen:  []string{schema_definition.AttributeAttachmentFile},
		AtLeastOneOf:  []string{schema_definition.AttributeAttachmentFile},
		ForceNew:      true,
		Optional:      true,
		Computed:      true,
	}

	resourceAttachmentSchema[schema_definition.AttributeAttachmentItemID] = &schema.Schema{
		Description: schema_definition.DescriptionItemIdentifier,
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
	d.Set(schema_definition.AttributeAttachmentItemID, split[1])
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
