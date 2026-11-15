package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	epschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SecureNoteResourceSchema() rsschema.Schema {
	attrs := itemBaseResourceAttributes()
	attrs[AttributeFavorite] = rsschema.BoolAttribute{Description: DescriptionFavorite, Optional: true, Computed: true}
	attrs[AttributeAttachments] = attachmentsResourceAttribute()

	return rsschema.Schema{
		Description: "Manages a secure note item.",
		Attributes:  attrs,
		Blocks: map[string]rsschema.Block{
			AttributeField: fieldResourceBlock(),
		},
	}
}

func SecureNoteDataSourceSchema() dsschema.Schema {
	attrs := itemBaseDataSourceAttributes()
	attrs[AttributeFavorite] = dsschema.BoolAttribute{Description: DescriptionFavorite, Computed: true}
	attrs[AttributeAttachments] = attachmentsDataSourceAttribute()

	return dsschema.Schema{
		MarkdownDescription: "Use this data source to get information on an existing secure note item. " +
			"The notes and fields are stored in Terraform plan and state. " +
			"To read a secure note item without persisting it, use the `bitwarden_item_secure_note` ephemeral resource.",
		Attributes: attrs,
	}
}

func SecureNoteEphemeralResourceSchema() epschema.Schema {
	attrs := itemBaseEphemeralAttributes()
	attrs[AttributeFavorite] = epschema.BoolAttribute{Description: DescriptionFavorite, Computed: true}
	attrs[AttributeAttachments] = attachmentsEphemeralAttribute()

	return epschema.Schema{
		MarkdownDescription: "Reads an existing Password Manager secure note item without persisting its data in plan or state files. " +
			"Unlike the `bitwarden_item_secure_note` data source, values are not stored in plan or state. " +
			"Use the result in an ephemeral context, such as a provider configuration or a write-only resource argument.",
		Attributes: attrs,
	}
}
