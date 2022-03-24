package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

func dataSourceItemSecureNote() *schema.Resource {
	dataSourceItemSecureNoteSchema := baseSchema(DataSource)

	return &schema.Resource{
		Description: "Use this data source to get (amongst other things) the content of a Bitwarden Secret Note for use in other resources.",
		ReadContext: dataSourceItemSecureNoteRead,
		Schema:      dataSourceItemSecureNoteSchema,
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
