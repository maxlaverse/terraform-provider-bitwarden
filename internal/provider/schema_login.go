package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func loginSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		attributeLoginPassword: {
			Description: descriptionLoginPassword,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeLoginUsername: {
			Description: descriptionLoginUsername,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeLoginTotp: {
			Description: descriptionLoginTotp,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		attributeLoginURIs: {
			Description: descriptionLoginUri,
			Type:        schema.TypeList,
			Elem:        uriElem(),
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   false,
		},
	}

	if schemaType == DataSource {
		base[attributeFilterURL] = &schema.Schema{
			Description: descriptionFilterURL,
			Type:        schema.TypeString,
			Optional:    true,
		}
	}
	return base
}

func uriElem() *schema.Resource {
	validMatchStr := []string{"default", "base_domain", "host", "start_with", "exact", "regexp", "never"}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			attributeLoginURIsMatch: {
				Description:      descriptionLoginUriMatch,
				Type:             schema.TypeString,
				Default:          validMatchStr[0],
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validMatchStr, false)),
				Optional:         true,
			},
			attributeLoginURIsValue: {
				Description: descriptionLoginUriValue,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
