package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
		Description: "Use this data source to get information on an existing secure note item.",
		Attributes:  attrs,
	}
}
