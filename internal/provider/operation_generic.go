package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceImporter(stateContext schema.StateContextFunc) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: stateContext,
	}
}
