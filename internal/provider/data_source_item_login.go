package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

func dataSourceItemLogin() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get (amongst other things) the username and password of a Bitwarden Login item for use in other resources.",

		ReadContext: dataSourceItemLoginRead,

		Schema: map[string]*schema.Schema{
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
			attributeLoginPassword: {
				Description: descriptionLoginPassword,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeLoginUsername: {
				Description: descriptionLoginUsername,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeLoginTotp: {
				Description: descriptionLoginTotp,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeName: {
				Description: descriptionName,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeNotes: {
				Description: descriptionNotes,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeObject: {
				Description: descriptionInternal,
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

func dataSourceItemLoginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get(attributeID).(string))
	err := d.Set(attributeObject, bitwarden.ObjectTypeItem)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set(attributeType, bitwarden.ItemTypeLogin)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectRead(ctx, d, meta)
}
