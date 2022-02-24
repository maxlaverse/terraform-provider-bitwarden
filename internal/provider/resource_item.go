package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceItemField() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			attributeFieldName: {
				Type:     schema.TypeString,
				Required: true,
			},
			attributeFieldText: {
				Type:     schema.TypeString,
				Optional: true,
			},
			attributeFieldBoolean: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			attributeFieldHidden: {
				Type:     schema.TypeString,
				Optional: true,
			},
			attributeFieldLinked: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}
