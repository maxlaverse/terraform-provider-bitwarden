package schema_definition

import (
	"context"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	epschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
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
	attrs[AttributeLoginPassword] = optionalComputedSensitiveString(DescriptionLoginPassword)
	attrs[AttributeLoginUsername] = optionalComputedSensitiveString(DescriptionLoginUsername)
	attrs[AttributeLoginTotp] = optionalComputedSensitiveString(DescriptionLoginTotp)
	attrs[AttributeFavorite] = optionalComputedBool(DescriptionFavorite)
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
		MarkdownDescription: "Use this data source to get information on an existing login item. " +
			"The login username, password, TOTP, notes, and fields are stored in Terraform plan and state. " +
			"To read a login item without persisting it, use the `bitwarden_item_login` ephemeral resource.",
		Attributes: attrs,
	}
}

func LoginEphemeralResourceSchema() epschema.Schema {
	attrs := itemBaseEphemeralAttributes()
	attrs[AttributeLoginPassword] = epschema.StringAttribute{Description: DescriptionLoginPassword, Computed: true, Sensitive: true}
	attrs[AttributeLoginUsername] = epschema.StringAttribute{Description: DescriptionLoginUsername, Computed: true, Sensitive: true}
	attrs[AttributeLoginTotp] = epschema.StringAttribute{Description: DescriptionLoginTotp, Computed: true, Sensitive: true}
	attrs[AttributeFavorite] = epschema.BoolAttribute{Description: DescriptionFavorite, Computed: true}
	attrs[AttributeLoginURIs] = uriEphemeralAttribute()
	attrs[AttributeAttachments] = attachmentsEphemeralAttribute()
	attrs[AttributeFilterURL] = epschema.StringAttribute{Description: DescriptionFilterURL, Optional: true}

	return epschema.Schema{
		MarkdownDescription: "Reads an existing Password Manager login item without persisting its data in plan or state files. " +
			"Unlike the `bitwarden_item_login` data source, values are not stored in plan or state. " +
			"Use the result in an ephemeral context, such as a provider configuration or a write-only resource argument.",
		Attributes: attrs,
	}
}
