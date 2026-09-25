package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SSHKeyResourceSchema() rsschema.Schema {
	attrs := itemBaseResourceAttributes()
	attrs[AttributeSSHKeyPrivateKey] = rsschema.StringAttribute{Description: DescriptionPrivateKey, Optional: true, Computed: true, Sensitive: true}
	attrs[AttributeSSHKeyPublicKey] = rsschema.StringAttribute{Description: DescriptionPublicKey, Optional: true, Computed: true, Sensitive: true}
	attrs[AttributeSSHKeyKeyFingerprint] = rsschema.StringAttribute{Description: DescriptionKeyFingerprint, Optional: true, Computed: true, Sensitive: true}

	return rsschema.Schema{
		Description: "Manages an SSH key item.",
		Attributes:  attrs,
		Blocks: map[string]rsschema.Block{
			AttributeField: fieldResourceBlock(),
		},
	}
}

func SSHKeyDataSourceSchema() dsschema.Schema {
	attrs := itemBaseDataSourceAttributes()
	attrs[AttributeSSHKeyPrivateKey] = dsschema.StringAttribute{Description: DescriptionPrivateKey, Computed: true, Sensitive: true}
	attrs[AttributeSSHKeyPublicKey] = dsschema.StringAttribute{Description: DescriptionPublicKey, Computed: true, Sensitive: true}
	attrs[AttributeSSHKeyKeyFingerprint] = dsschema.StringAttribute{Description: DescriptionKeyFingerprint, Computed: true, Sensitive: true}

	return dsschema.Schema{
		Description: "Use this data source to get information on an existing SSH key item.",
		Attributes:  attrs,
	}
}
