package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func resourceAttachment() *schema.Resource {
	resourceAttachmentSchema := schema_definition.AttachmentSchema()
	resourceAttachmentSchema[schema_definition.AttributeAttachmentFile] = &schema.Schema{
		Description:   schema_definition.DescriptionItemAttachmentFile,
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{schema_definition.AttributeAttachmentContent},
		AtLeastOneOf:  []string{schema_definition.AttributeAttachmentContent},
		ForceNew:      true,
	}

	resourceAttachmentSchema[schema_definition.AttributeAttachmentContent] = &schema.Schema{
		Description:   schema_definition.DescriptionItemAttachmentFile,
		Type:          schema.TypeString,
		Optional:      true,
		RequiredWith:  []string{schema_definition.AttributeAttachmentContent},
		ConflictsWith: []string{schema_definition.AttributeAttachmentFile},
		AtLeastOneOf:  []string{schema_definition.AttributeAttachmentFile},
		ForceNew:      true,
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
	resourceAttachmentSchema[schema_definition.AttributeAttachmentDataHash] = &schema.Schema{
		Description:  schema_definition.DescriptionItemAttachmentDataHash,
		Type:         schema.TypeString,
		Optional:     true,
		AtLeastOneOf: []string{schema_definition.AttributeAttachmentContent, schema_definition.AttributeAttachmentFile},
		ForceNew:     true,
		Computed:     true,
	}
	resourceAttachmentSchema[schema_definition.AttributeAttachmentItemID] = &schema.Schema{
		Description: schema_definition.DescriptionItemIdentifier,
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}

	return &schema.Resource{
		Description: "Manages an item attachment.",
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {

			// Commented code below was used to learn how values would come in through CustomizeDiff.
			// 1) List every key that has a diff (state vs plan):
			changed := d.GetChangedKeysPrefix("")
			tflog.Trace(ctx, "🔍 CustomizeDiff: changed keys", map[string]interface{}{
				"keys": changed,
			})

			// 2) For each changed key, dump old vs new:
			for _, key := range changed {
				oldVal, newVal := d.GetChange(key)
				tflog.Trace(ctx, fmt.Sprintf("🔄 diff for %q", key), map[string]interface{}{
					"old": oldVal,
					"new": newVal,
				})
			}
			schemaKeys := make([]string, 0, len(resourceAttachmentSchema))
			for key := range resourceAttachmentSchema {
				schemaKeys = append(schemaKeys, key)
			}
			// 3) Show the “current” config values (what’s in HCL or interpolated):
			//    Use GetOk to fetch any key that exists in config/new-diff.
			for _, key := range schemaKeys {
				if v, ok := d.GetOk(key); ok {
					tflog.Trace(ctx, fmt.Sprintf("🏷️ config \"%s\"", key), map[string]interface{}{
						"value": v,
					})
				} else {
					tflog.Trace(ctx, fmt.Sprintf("🏷️ config \"%s\" (absent)", key), nil)
				}
			}

			tflog.Trace(ctx, "✅ CustomizeDiff dump complete", nil)
			//confFilePath, fileSpecified := d.GetOk(schema_definition.AttributeAttachmentFile)
			filePathOldValue, filePathNewValue := d.GetChange(schema_definition.AttributeAttachmentFile)
			tflog.Trace(ctx, fmt.Sprintf("🔄🔄 FORCED for %q", schema_definition.AttributeAttachmentFile), map[string]interface{}{
				"old": filePathOldValue,
				"new": filePathNewValue,
			})
			//confContent, contentSpecified := d.GetOk(schema_definition.AttributeAttachmentContent)
			contentOldValue, contentNewValue := d.GetChange(schema_definition.AttributeAttachmentContent)
			tflog.Trace(ctx, fmt.Sprintf("🔄🔄 FORCED for %q", schema_definition.AttributeAttachmentContent), map[string]interface{}{
				"old": contentOldValue,
				"new": contentNewValue,
			})
			//confHash, hashSpecified := d.GetOk(schema_definition.AttributeAttachmentDataHash)
			hashOldValue, hashNewValue := d.GetChange(schema_definition.AttributeAttachmentDataHash)
			tflog.Trace(ctx, fmt.Sprintf("🔄🔄 FORCED diff for %q", schema_definition.AttributeAttachmentDataHash), map[string]interface{}{
				"old": hashOldValue,
				"new": hashNewValue,
			})
			// if there is a change, return as the update will happen without additional checking.
			if filePathOldValue != filePathNewValue {
				return nil
			}
			if contentOldValue != contentNewValue {
				return nil
			}
			if hashOldValue != hashNewValue {
				return nil
			}

			// if we got to here, check if the file on disk has been updated. file specified as content does not need
			// this check, as it will be different and auto trigger an update.
			if filePathOldValue != "" && hashOldValue != "" {
				fileOnDiskHash := fileHash(filePathOldValue)
				tflog.Debug(ctx, fmt.Sprintf("✅✅ FORCED fileOnDiskHash:"), map[string]interface{}{
					"fileOnDiskHash": fileOnDiskHash,
					"hashOldValue":   hashOldValue,
				})
				if fileOnDiskHash != hashOldValue {
					// hashes are different, force new.
					setNewErr := d.SetNew(schema_definition.AttributeAttachmentDataHash, fileOnDiskHash)
					if setNewErr != nil {
						tflog.Warn(ctx, fmt.Sprintf("SetNew failed for %q: %s", schema_definition.AttributeAttachmentDataHash, setNewErr))
					}
					err := d.ForceNew(schema_definition.AttributeAttachmentDataHash)
					if err != nil {
						tflog.Warn(ctx, fmt.Sprintf("ForceNew failed for %q: %s", schema_definition.AttributeAttachmentDataHash, err))
					}
				}
			}
			return nil
		},
		CreateContext: withPasswordManager(opAttachmentCreate),
		ReadContext:   withPasswordManager(opAttachmentReadIgnoreMissing),
		DeleteContext: withPasswordManager(opAttachmentDelete),
		Importer:      resourceImporter(opAttachmentImport),

		Schema: resourceAttachmentSchema,
	}
}
