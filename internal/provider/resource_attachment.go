package provider

import (
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

		CreateContext: withPasswordManager(opAttachmentCreate),
		ReadContext:   withPasswordManager(opAttachmentReadIgnoreMissing),
		DeleteContext: withPasswordManager(opAttachmentDelete),
		Importer:      resourceImporter(opAttachmentImport),

		Schema: resourceAttachmentSchema,
	}
}
