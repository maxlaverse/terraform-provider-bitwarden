package schema_definition

const (
	// Data-source and Resource field attributes
	AttributeAttachments                   = "attachments"
	AttributeCollectionIDs                 = "collection_ids"
	AttributeCollectionMemberOrgMemberId   = "org_member_id"
	AttributeCollectionMemberReadOnly      = "read_only"
	AttributeCollectionMemberHidePasswords = "hide_passwords"
	AttributeCreationDate                  = "creation_date"
	AttributeDeletedDate                   = "deleted_date"
	AttributeID                            = "id"
	AttributeFavorite                      = "favorite"
	AttributeField                         = "field"
	AttributeFieldName                     = "name"
	AttributeFieldBoolean                  = "boolean"
	AttributeFieldHidden                   = "hidden"
	AttributeFieldLinked                   = "linked"
	AttributeFieldText                     = "text"
	AttributeFolderID                      = "folder_id"
	AttributeAttachmentContent             = "content"
	AttributeAttachmentItemID              = "item_id"
	AttributeAttachmentFile                = "file"
	AttributeAttachmentSize                = "size"
	AttributeAttachmentSizeName            = "size_name"
	AttributeAttachmentFileName            = "file_name"
	AttributeAttachmentURL                 = "url"
	AttributeEmail                         = "email"
	AttributeFilterCollectionId            = "filter_collection_id"
	AttributeFilterFolderID                = "filter_folder_id"
	AttributeFilterOrganizationID          = "filter_organization_id"
	AttributeFilterSearch                  = "search"
	AttributeFilterURL                     = "filter_url"
	AttributeLoginPassword                 = "password"
	AttributeLoginUsername                 = "username"
	AttributeLoginURIs                     = "uri"
	AttributeLoginURIsMatch                = "match"
	AttributeLoginURIsValue                = "value"
	AttributeLoginTotp                     = "totp"
	AttributeMember                        = "member"
	AttributeName                          = "name"
	AttributeNotes                         = "notes"
	AttributeOrganizationID                = "organization_id"
	AttributeReprompt                      = "reprompt"
	AttributeRevisionDate                  = "revision_date"

	// Secret specific attributes
	AttributeKey       = "key"
	AttributeNote      = "note"
	AttributeProjectID = "project_id"
	AttributeValue     = "value"

	// Data-source and Resource field descriptions
	DescriptionAttachments                   = "List of item attachments."
	DescriptionCollectionIDs                 = "Identifier of the collections the item belongs to."
	DescriptionCollectionMember              = "[Experimental] Member of a collection."
	DescriptionCollectionMemberReadOnly      = "[Experimental] Read/Write permissions."
	DescriptionCollectionMemberHidePasswords = "[Experimental] Hide passwords."
	DescriptionCreationDate                  = "Date the item was created."
	DescriptionDeletedDate                   = "Date the item was deleted."
	DescriptionEmail                         = "User email."
	DescriptionFavorite                      = "Mark as a Favorite to have item appear at the top of your Vault in the UI."
	DescriptionField                         = "Extra fields."
	DescriptionFieldBoolean                  = "Value of a boolean field."
	DescriptionFieldHidden                   = "Value of a hidden text field."
	DescriptionFieldLinked                   = "Value of a linked field."
	DescriptionFieldName                     = "Name of the field."
	DescriptionFieldText                     = "Value of a text field."
	DescriptionFilterCollectionID            = "Filter search results by collection ID."
	DescriptionFilterFolderID                = "Filter search results by folder ID."
	DescriptionFilterOrganizationID          = "Filter search results by organization ID."
	DescriptionFilterSearch                  = "Search items matching the search string."
	DescriptionFilterURL                     = "Filter search results by URL."
	DescriptionFolderID                      = "Identifier of the folder."
	DescriptionIdentifier                    = "Identifier."
	DescriptionItemIdentifier                = "Identifier of the item the attachment belongs to"
	DescriptionItemAttachmentContent         = "Content of the attachment"
	DescriptionItemAttachmentFile            = "Path to the content of the attachment."
	DescriptionItemAttachmentFileName        = "File name"
	DescriptionItemAttachmentSize            = "Size in bytes"
	DescriptionItemAttachmentSizeName        = "Size as string"
	DescriptionItemAttachmentURL             = "URL"
	DescriptionLoginPassword                 = "Login password."
	DescriptionLoginUri                      = "URI."
	DescriptionLoginUriMatch                 = "URI Match"
	DescriptionLoginUriValue                 = "URI Value"
	DescriptionLoginTotp                     = "Verification code."
	DescriptionLoginUsername                 = "Login username."
	DescriptionName                          = "Name."
	DescriptionNotes                         = "Notes."
	DescriptionOrganizationID                = "Identifier of the organization."
	DescriptionReprompt                      = "Require master password “re-prompt” when displaying secret in the UI."
	DescriptionRevisionDate                  = "Last time the item was updated."

	// Secret specific attributes
	DescriptionValue     = "Value."
	DescriptionNote      = "Note."
	DescriptionProjectID = "Identifier of the project."

	// Provider field attributes
	AttributeAccessToken                = "access_token"
	AttributeClientID                   = "client_id"
	AttributeClientSecret               = "client_secret"
	AttributeProviderEmail              = "email"
	AttributeMasterPassword             = "master_password"
	AttributeServer                     = "server"
	AttributeSessionKey                 = "session_key"
	AttributeVaultPath                  = "vault_path"
	AttributeExtraCACertsPath           = "extra_ca_certs"
	AttributeExperimental               = "experimental"
	AttributeExperimentalEmbeddedClient = "embedded_client"

	// Provider field descriptions
	DescriptionAccessToken                = "Machine Account Access Token (env: `BWS_ACCESS_TOKEN`))."
	DescriptionClientSecret               = "Client Secret (env: `BW_CLIENTSECRET`). Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	DescriptionClientID                   = "Client ID (env: `BW_CLIENTID`)"
	DescriptionProviderEmail              = "Login Email of the Vault (env: `BW_EMAIL`)."
	DescriptionMasterPassword             = "Master password of the Vault (env: `BW_PASSWORD`). Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	DescriptionServer                     = "Bitwarden Server URL (default: `https://vault.bitwarden.com`, env: `BW_URL`)."
	DescriptionSessionKey                 = "A Bitwarden Session Key (env: `BW_SESSION`)"
	DescriptionVaultPath                  = "Alternative directory for storing the Vault locally (default: `.bitwarden/`, env: `BITWARDENCLI_APPDATA_DIR`)."
	DescriptionExtraCACertsPath           = "Extends the well known 'root' CAs (like VeriSign) with the extra certificates in file (env: `NODE_EXTRA_CA_CERTS`)."
	DescriptionExperimental               = "Enable experimental features."
	DescriptionExperimentalEmbeddedClient = "Use the embedded client instead of an external binary."
)
