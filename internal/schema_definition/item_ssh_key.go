package schema_definition

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	epschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func SSHKeyResourceSchema() rsschema.Schema {
	attrs := itemBaseResourceAttributes()
	attrs[AttributeSSHKeyPrivateKey] = optionalComputedSensitiveString(DescriptionPrivateKey)
	attrs[AttributeSSHKeyPublicKey] = optionalComputedSensitiveString(DescriptionPublicKey)
	attrs[AttributeSSHKeyKeyFingerprint] = optionalComputedSensitiveString(DescriptionKeyFingerprint)

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
		MarkdownDescription: "Use this data source to get information on an existing SSH key item. " +
			"The private key, public key, fingerprint, notes, and fields are stored in Terraform plan and state. " +
			"To read an SSH key item without persisting it, use the `bitwarden_item_ssh_key` ephemeral resource.",
		Attributes: attrs,
	}
}

func SSHKeyEphemeralResourceSchema() epschema.Schema {
	attrs := itemBaseEphemeralAttributes()
	attrs[AttributeSSHKeyPrivateKey] = epschema.StringAttribute{Description: DescriptionPrivateKey, Computed: true, Sensitive: true}
	attrs[AttributeSSHKeyPublicKey] = epschema.StringAttribute{Description: DescriptionPublicKey, Computed: true, Sensitive: true}
	attrs[AttributeSSHKeyKeyFingerprint] = epschema.StringAttribute{Description: DescriptionKeyFingerprint, Computed: true, Sensitive: true}

	return epschema.Schema{
		MarkdownDescription: "Reads an existing Password Manager SSH key item without persisting its data in plan or state files. " +
			"Unlike the `bitwarden_item_ssh_key` data source, values are not stored in plan or state. " +
			"Use the result in an ephemeral context, such as a provider configuration or a write-only resource argument.",
		Attributes: attrs,
	}
}
