package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func resourceItemLogin() *schema.Resource {
	return &schema.Resource{
		Description: "Use this resource to set (amongst other things) the username and password of a Bitwarden Login item.",

		CreateContext: resourceItemLoginCreate,
		ReadContext:   objectRead,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,

		Schema: map[string]*schema.Schema{
			attributeCollectionIDs: {
				Description: descriptionCollectionIDs,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			attributeID: {
				Description: descriptionIdentifier,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeFavorite: {
				Description: descriptionFavorite,
				Type:        schema.TypeBool,
				Optional:    true,
			},
			attributeField: {
				Description: descriptionField,
				Type:        schema.TypeList,
				Elem:        resourceItemField(),
				Optional:    true,
			},
			attributeFolderID: {
				Description: descriptionFolderID,
				Type:        schema.TypeString,
				Optional:    true,
			},
			attributeLoginPassword: {
				Description: descriptionLoginPassword,
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			attributeLoginUsername: {
				Description: descriptionLoginUsername,
				Type:        schema.TypeString,
				Optional:    true,
			},
			attributeLoginTotp: {
				Description: descriptionLoginTotp,
				Type:        schema.TypeString,
				Optional:    true,
			},
			attributeName: {
				Description: descriptionName,
				Type:        schema.TypeString,
				Required:    true,
			},
			attributeNotes: {
				Description: descriptionNotes,
				Type:        schema.TypeString,
				Optional:    true,
			},
			attributeObject: {
				Description: descriptionInternal,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeOrganizationID: {
				Description: descriptionOrganizationID,
				Type:        schema.TypeString,
				Optional:    true,
			},
			attributeReprompt: {
				Description: descriptionReprompt,
				Type:        schema.TypeBool,
				Optional:    true,
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

func resourceItemLoginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(attributeObject, bw.ObjectTypeItem)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set(attributeType, bw.ItemTypeLogin)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectCreate(ctx, d, meta)
}
