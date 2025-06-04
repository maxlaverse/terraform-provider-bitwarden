package schema_definition

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

type URIMatchStr string

const (
	URIMatchDefaultStr    URIMatchStr = "default"
	URIMatchBaseDomainStr URIMatchStr = "base_domain"
	URIMatchHostStr       URIMatchStr = "host"
	URIMatchStartWithStr  URIMatchStr = "start_with"
	URIMatchExactStr      URIMatchStr = "exact"
	URIMatchRegExpStr     URIMatchStr = "regexp"
	URIMatchNeverStr      URIMatchStr = "never"
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
		AttributeFavorite: {
			Description: DescriptionFavorite,
			Type:        schema.TypeBool,
			Computed:    schemaType == DataSource,
			Optional:    schemaType == Resource,
		},
		AttributeAttachments: {
			Description: DescriptionAttachments,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: AttachmentSchema(),
			},
			Computed: true,
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
	validMatchStr := []string{
		string(URIMatchDefaultStr),
		string(URIMatchBaseDomainStr),
		string(URIMatchHostStr),
		string(URIMatchStartWithStr),
		string(URIMatchExactStr),
		string(URIMatchRegExpStr),
		string(URIMatchNeverStr),
	}

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

func IntMatchToStr(ctx context.Context, match *models.URIMatch) URIMatchStr {
	if match == nil {
		return URIMatchDefaultStr
	}

	switch *match {
	case models.URIMatchBaseDomain:
		return URIMatchBaseDomainStr
	case models.URIMatchHost:
		return URIMatchHostStr
	case models.URIMatchStartWith:
		return URIMatchStartWithStr
	case models.URIMatchExact:
		return URIMatchExactStr
	case models.URIMatchRegExp:
		return URIMatchRegExpStr
	case models.URIMatchNever:
		return URIMatchNeverStr
	default:
		tflog.Warn(ctx, "unsupported integer value for URI match - Falling back to default", map[string]interface{}{"match": *match})
		return URIMatchDefaultStr
	}
}

func StrMatchToInt(ctx context.Context, match string) *models.URIMatch {
	var v models.URIMatch
	switch match {
	case string(URIMatchDefaultStr):
		return nil
	case string(URIMatchBaseDomainStr):
		v = models.URIMatchBaseDomain
	case string(URIMatchHostStr):
		v = models.URIMatchHost
	case string(URIMatchStartWithStr):
		v = models.URIMatchStartWith
	case string(URIMatchExactStr):
		v = models.URIMatchExact
	case string(URIMatchRegExpStr):
		v = models.URIMatchRegExp
	case string(URIMatchNeverStr):
		v = models.URIMatchNever
	default:
		tflog.Warn(ctx, "unsupported string value for URI match - Falling back to default", map[string]interface{}{"match": match})
		return nil
	}
	return &v
}
