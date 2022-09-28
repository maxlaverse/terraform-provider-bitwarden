package provider

const (
	// Datasource and Resource field attributes
	attributeCollectionIDs  = "collection_ids"
	attributeID             = "id"
	attributeFavorite       = "favorite"
	attributeField          = "field"
	attributeFieldName      = "name"
	attributeFieldBoolean   = "boolean"
	attributeFieldHidden    = "hidden"
	attributeFieldLinked    = "linked"
	attributeFieldText      = "text"
	attributeFolderID       = "folder_id"
	attributeLoginPassword  = "password"
	attributeLoginUsername  = "username"
	attributeLoginTotp      = "totp"
	attributeName           = "name"
	attributeNotes          = "notes"
	attributeObject         = "object"
	attributeOrganizationID = "organization_id"
	attributeReprompt       = "reprompt"
	attributeRevisionDate   = "revision_date"
	attributeType           = "type"

	// Datasource and Resource field descriptions
	descriptionCollectionIDs  = "Identifier of the collections the item belongs to."
	descriptionFavorite       = "Mark as a Favorite to have item appear at the top of your Vault in the UI."
	descriptionField          = "Extra fields."
	descriptionFieldBoolean   = "Value of a boolean field."
	descriptionFieldHidden    = "Value of a hidden text field."
	descriptionFieldLinked    = "Value of a linked field."
	descriptionFieldName      = "Name of the field."
	descriptionFieldText      = "Value of a text field."
	descriptionFolderID       = "Identifier of the folder."
	descriptionIdentifier     = "Identifier."
	descriptionInternal       = "INTERNAL USE" // TODO: Manage to hide this from the users
	descriptionLoginPassword  = "Login password."
	descriptionLoginTotp      = "Verification code."
	descriptionLoginUsername  = "Login username."
	descriptionName           = "Name."
	descriptionNotes          = "Notes."
	descriptionOrganizationID = "Identifier of the organization."
	descriptionReprompt       = "Require master password “re-prompt” when displaying secret in the UI."
	descriptionRevisionDate   = "Last time the item was updated."

	// Provider field attributes
	attributeClientID       = "client_id"
	attributeClientSecret   = "client_secret"
	attributeEmail          = "email"
	attributeMasterPassword = "master_password"
	attributeServer         = "server"
	attributeSessionKey     = "session_key"
	attributeVaultPath      = "vault_path"

	// Provider field descriptions
	descriptionClientSecret   = "Client Secret. Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionClientID       = "Client ID."
	descriptionEmail          = "Login Email of the Vault."
	descriptionMasterPassword = "Master password of the Vault. Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionServer         = "Bitwarden Server URL (default: https://vault.bitwarden.com)."
	descriptionSessionKey     = "A Bitwarden Session Key."
	descriptionVaultPath      = "Alternative directory for storing the Vault locally."
)
