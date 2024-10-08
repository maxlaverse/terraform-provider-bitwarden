package provider

const (
	// Data-source and Resource field attributes
	attributeAttachments          = "attachments"
	attributeCollectionIDs        = "collection_ids"
	attributeCreationDate         = "creation_date"
	attributeDeletedDate          = "deleted_date"
	attributeID                   = "id"
	attributeFavorite             = "favorite"
	attributeField                = "field"
	attributeFieldName            = "name"
	attributeFieldBoolean         = "boolean"
	attributeFieldHidden          = "hidden"
	attributeFieldLinked          = "linked"
	attributeFieldText            = "text"
	attributeFilterValues         = "values"
	attributeFolderID             = "folder_id"
	attributeAttachmentContent    = "content"
	attributeAttachmentItemID     = "item_id"
	attributeAttachmentFile       = "file"
	attributeAttachmentSize       = "size"
	attributeAttachmentSizeName   = "size_name"
	attributeAttachmentFileName   = "file_name"
	attributeAttachmentURL        = "url"
	attributeFilterCollectionId   = "filter_collection_id"
	attributeFilterFolderID       = "filter_folder_id"
	attributeFilterOrganizationID = "filter_organization_id"
	attributeFilterSearch         = "search"
	attributeFilterURL            = "filter_url"
	attributeLoginPassword        = "password"
	attributeLoginUsername        = "username"
	attributeLoginURIs            = "uri"
	attributeLoginURIsMatch       = "match"
	attributeLoginURIsValue       = "value"
	attributeLoginTotp            = "totp"
	attributeName                 = "name"
	attributeNotes                = "notes"
	attributeObject               = "object"
	attributeOrganizationID       = "organization_id"
	attributeReprompt             = "reprompt"
	attributeRevisionDate         = "revision_date"
	attributeType                 = "type"

	// Secret specific attributes
	attributeKey       = "key"
	attributeNote      = "note"
	attributeProjectID = "project_id"
	attributeValue     = "value"

	// Data-source and Resource field descriptions
	descriptionAttachments            = "List of item attachments."
	descriptionCollectionIDs          = "Identifier of the collections the item belongs to."
	descriptionCreationDate           = "Date the item was created."
	descriptionDeletedDate            = "Date the item was deleted."
	descriptionFavorite               = "Mark as a Favorite to have item appear at the top of your Vault in the UI."
	descriptionField                  = "Extra fields."
	descriptionFieldBoolean           = "Value of a boolean field."
	descriptionFieldHidden            = "Value of a hidden text field."
	descriptionFieldLinked            = "Value of a linked field."
	descriptionFieldName              = "Name of the field."
	descriptionFieldText              = "Value of a text field."
	descriptionFilterCollectionID     = "Filter search results by collection ID."
	descriptionFilterFolderID         = "Filter search results by folder ID."
	descriptionFilterOrganizationID   = "Filter search results by organization ID."
	descriptionFilterSearch           = "Search items matching the search string."
	descriptionFilterURL              = "Filter search results by URL."
	descriptionFolderID               = "Identifier of the folder."
	descriptionIdentifier             = "Identifier."
	descriptionInternal               = "INTERNAL USE"
	descriptionItemIdentifier         = "Identifier of the item the attachment belongs to"
	descriptionItemAttachmentContent  = "Content of the attachment"
	descriptionItemAttachmentFile     = "Path to the content of the attachment."
	descriptionItemAttachmentFileName = "File name"
	descriptionItemAttachmentSize     = "Size in bytes"
	descriptionItemAttachmentSizeName = "Size as string"
	descriptionItemAttachmentURL      = "URL"
	descriptionLoginPassword          = "Login password."
	descriptionLoginUri               = "URI."
	descriptionLoginUriMatch          = "URI Match"
	descriptionLoginUriValue          = "URI Value"
	descriptionLoginTotp              = "Verification code."
	descriptionLoginUsername          = "Login username."
	descriptionName                   = "Name."
	descriptionNotes                  = "Notes."
	descriptionOrganizationID         = "Identifier of the organization."
	descriptionReprompt               = "Require master password “re-prompt” when displaying secret in the UI."
	descriptionRevisionDate           = "Last time the item was updated."

	// Secret specific attributes
	descriptionValue     = "Value."
	descriptionNote      = "Note."
	descriptionProjectID = "Identifier of the project."

	// Provider field attributes
	attributeAccessToken                = "access_token"
	attributeClientID                   = "client_id"
	attributeClientSecret               = "client_secret"
	attributeEmail                      = "email"
	attributeMasterPassword             = "master_password"
	attributeServer                     = "server"
	attributeSessionKey                 = "session_key"
	attributeVaultPath                  = "vault_path"
	attributeExtraCACertsPath           = "extra_ca_certs"
	attributeExperimental               = "experimental"
	attributeExperimentalEmbeddedClient = "embedded_client"

	// Provider field descriptions
	descriptionAccessToken                = "Machine Account Access Token (env: `BWS_ACCESS_TOKEN`))."
	descriptionClientSecret               = "Client Secret (env: `BW_CLIENTSECRET`). Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionClientID                   = "Client ID (env: `BW_CLIENTID`)"
	descriptionEmail                      = "Login Email of the Vault (env: `BW_EMAIL`)."
	descriptionMasterPassword             = "Master password of the Vault (env: `BW_PASSWORD`). Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionServer                     = "Bitwarden Server URL (default: `https://vault.bitwarden.com`, env: `BW_URL`)."
	descriptionSessionKey                 = "A Bitwarden Session Key (env: `BW_SESSION`)"
	descriptionVaultPath                  = "Alternative directory for storing the Vault locally (default: `.bitwarden/`, env: `BITWARDENCLI_APPDATA_DIR`)."
	descriptionExtraCACertsPath           = "Extends the well known 'root' CAs (like VeriSign) with the extra certificates in file (env: `NODE_EXTRA_CA_CERTS`)."
	descriptionExperimental               = "Enable experimental features."
	descriptionExperimentalEmbeddedClient = "Use the embedded client instead of an external binary."
)
