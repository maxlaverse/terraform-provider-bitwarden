package schema_definition

import (
	"context"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

// ValidURIMatchStrings returns the list of accepted string values for a login
// URI match strategy.
func ValidURIMatchStrings() []string {
	return []string{
		string(URIMatchDefaultStr),
		string(URIMatchBaseDomainStr),
		string(URIMatchHostStr),
		string(URIMatchStartWithStr),
		string(URIMatchExactStr),
		string(URIMatchRegExpStr),
		string(URIMatchNeverStr),
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

func LoginResourceSchema() rsschema.Schema {
	attrs := itemBaseResourceAttributes()
	attrs[AttributeLoginPassword] = rsschema.StringAttribute{Description: DescriptionLoginPassword, Optional: true, Computed: true, Sensitive: true}
	attrs[AttributeLoginUsername] = rsschema.StringAttribute{Description: DescriptionLoginUsername, Optional: true, Computed: true, Sensitive: true}
	attrs[AttributeLoginTotp] = rsschema.StringAttribute{Description: DescriptionLoginTotp, Optional: true, Computed: true, Sensitive: true}
	attrs[AttributeFavorite] = rsschema.BoolAttribute{Description: DescriptionFavorite, Optional: true, Computed: true}
	attrs[AttributeAttachments] = attachmentsResourceAttribute()

	return rsschema.Schema{
		Description: "Manages a login item.",
		Attributes:  attrs,
		Blocks: map[string]rsschema.Block{
			AttributeField:     fieldResourceBlock(),
			AttributeLoginURIs: uriResourceBlock(),
		},
	}
}

func LoginDataSourceSchema() dsschema.Schema {
	attrs := itemBaseDataSourceAttributes()
	attrs[AttributeLoginPassword] = dsschema.StringAttribute{Description: DescriptionLoginPassword, Computed: true, Sensitive: true}
	attrs[AttributeLoginUsername] = dsschema.StringAttribute{Description: DescriptionLoginUsername, Computed: true, Sensitive: true}
	attrs[AttributeLoginTotp] = dsschema.StringAttribute{Description: DescriptionLoginTotp, Computed: true, Sensitive: true}
	attrs[AttributeFavorite] = dsschema.BoolAttribute{Description: DescriptionFavorite, Computed: true}
	attrs[AttributeLoginURIs] = uriDataSourceAttribute()
	attrs[AttributeAttachments] = attachmentsDataSourceAttribute()
	attrs[AttributeFilterURL] = dsschema.StringAttribute{Description: DescriptionFilterURL, Optional: true}

	return dsschema.Schema{
		Description: "Use this data source to get information on an existing login item.",
		Attributes:  attrs,
	}
}
