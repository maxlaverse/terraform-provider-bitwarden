package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

func resourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: "Use this resource to create (amongst other things) a folder in Bitwarden, for storing other resources into.",

		CreateContext: resourceFolderCreate,
		ReadContext:   objectRead,
		UpdateContext: objectUpdate,
		DeleteContext: objectDelete,

		Schema: map[string]*schema.Schema{
			attributeID: {
				Description: descriptionIdentifier,
				Type:        schema.TypeString,
				Computed:    true,
			},
			attributeName: {
				Description: descriptionName,
				Type:        schema.TypeString,
				Required:    true,
			},
			attributeObject: {
				Description: descriptionInternal,
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(attributeObject, bitwarden.ObjectTypeFolder)
	if err != nil {
		return diag.FromErr(err)
	}
	return objectCreate(ctx, d, meta)
}
