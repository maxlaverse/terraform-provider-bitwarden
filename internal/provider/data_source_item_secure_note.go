package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceItemSecureNote() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get (amongst other things) the content of a Bitwarden Secret Note for use in other resources.",

		ReadContext: dataSourceItemSecureNoteRead,

		Schema: map[string]*schema.Schema{
			attributeCollectionIDs: {
				Description: descriptionCollectionIDs,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			attributeID: {
				Description: descriptionIdentifier,
				Type:        schema.TypeString,
				Required:    true,
			},
			attributeFavorite: {
				Description: descriptionFavorite,
				Type:        schema.TypeBool,
				Computed:    true,
			},
			attributeField: {
				Description: descriptionField,
				Type:        schema.TypeList,
				Elem:        resourceItemField(),
				Computed:    true,
			},
			attributeFolderID: {
				Description: descriptionFolderID,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeName: {
				Description: descriptionName,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeNotes: {
				Type:        schema.TypeString,
				Description: descriptionNotes,
				Computed:    true,
			},
			attributeObject: {
				Description: descriptionInternal,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeOrganizationID: {
				Description: descriptionOrganizationID,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeReprompt: {
				Description: descriptionReprompt,
				Type:        schema.TypeBool,
				Computed:    true,
			},
			attributeRevisionDate: {
				Description: descriptionRevisionDate,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeType: {
				Description: descriptionInternal,
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceItemSecureNoteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get(attributeID).(string))
	err := d.Set(attributeObject, bw.ObjectTypeItem)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set(attributeType, bw.ItemTypeLogin)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectRead(ctx, d, meta)
}
