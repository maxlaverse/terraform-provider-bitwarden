package schema_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func LoginSchema(schemaType schemaTypeEnum) map[string]*schema.Schema {
	base := map[string]*schema.Schema{
		AttributeLoginPassword: {
			Description: DescriptionLoginPassword,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		AttributeLoginUsername: {
			Description: DescriptionLoginUsername,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		AttributeLoginTotp: {
			Description: DescriptionLoginTotp,
			Type:        schema.TypeString,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   true,
		},
		AttributeLoginURIs: {
			Description: DescriptionLoginUri,
			Type:        schema.TypeList,
			Elem:        uriElem(),
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
			Sensitive:   false,
		},
	}

	if schemaType == DataSource {
		base[AttributeFilterURL] = &schema.Schema{
			Description: DescriptionFilterURL,
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
			AttributeLoginURIsMatch: {
				Description:      DescriptionLoginUriMatch,
				Type:             schema.TypeString,
				Default:          validMatchStr[0],
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validMatchStr, false)),
				Optional:         true,
			},
			AttributeLoginURIsValue: {
				Description: DescriptionLoginUriValue,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
