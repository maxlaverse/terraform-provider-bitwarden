package provider

const (
	// Datasource and Resource attribtues
	attributeID            = "id"
	attributeFavorite      = "favorite"
	attributeField         = "field"
	attributeFieldName     = "name"
	attributeFieldBoolean  = "boolean"
	attributeFieldHidden   = "hidden"
	attributeFieldLinked   = "linked"
	attributeFieldText     = "text"
	attributeFolderID      = "folder_id"
	attributeLoginPassword = "password"
	attributeLoginUsername = "username"
	attributeLoginTotp     = "totp"
	attributeName          = "name"
	attributeNotes         = "notes"
	attributeObject        = "object"
	attributeReprompt      = "reprompt"
	attributeRevisionDate  = "revision_date"
	attributeType          = "type"

	descriptionFavorite      = "Mark as a Favorite to have item appear at the top of your Vault in the UI."
	descriptionField         = "Extra fields."
	descriptionFolderID      = "Identifier of the folder."
	descriptionIdentifier    = "Identifier."
	descriptionInternal      = "INTERNAL USE" // TODO: Manage to hide this from the users
	descriptionLoginPassword = "Login password."
	descriptionLoginTotp     = "Verification code."
	descriptionLoginUsername = "Login username."
	descriptionName          = "Name."
	descriptionNotes         = "Notes."
	descriptionReprompt      = "Require master password “re-prompt” when displaying secret in the UI."
	descriptionRevisionDate  = "Last time the item was updated."

	// Provider attributes
	attributeClientID       = "client_id"
	attributeClientSecret   = "client_secret"
	attributeEmail          = "email"
	attributeMasterPassword = "master_password"
	attributeServer         = "server"
	attributeVaultPath      = "vault_path"

	descriptionClientSecret   = "Client Secret. Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionClientID       = "Client ID."
	descriptionEmail          = "Login Email of the Vault."
	descriptionMasterPassword = "Master password of the Vault. Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionServer         = "Bitwarden server URL (default: https://vault.bitwarden.com)."
	descriptionVaultPath      = "Alternative directory for storing the Vault locally."
)
