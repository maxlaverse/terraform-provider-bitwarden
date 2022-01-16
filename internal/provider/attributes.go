package provider

const (
	// Datasource and Resource attribtues
	attributeID            = "id"
	attributeFolderID      = "folder_id"
	attributeLoginPassword = "password"
	attributeLoginUsername = "username"
	attributeLoginTotp     = "totp"
	attributeName          = "name"
	attributeNotes         = "notes"
	attributeObject        = "object"
	attributeType          = "type"

	descriptionIdentifier    = "Identifier."
	descriptionLoginTotp     = "Verification code."
	descriptionLoginUsername = "Login username."
	descriptionLoginPassword = "Login password."
	descriptionFolderID      = "Identifier of the folder."
	descriptionName          = "Name."
	descriptionNotes         = "Notes."
	descriptionInternal      = "INTERNAL USE" // TODO: Manage to hide this from the users

	// Provider attributes
	attributeMasterPassword = "master_password"
	attributeClientID       = "client_id"
	attributeClientSecret   = "client_secret"
	attributeServer         = "server"
	attributeEmail          = "email"
	attributeVaultPath      = "vault_path"

	descriptionMasterPassword = "Master password of the Vault. Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionClientID       = "Client API (required when authenticating with the official Bitwarden instances)."
	descriptionClientSecret   = "Client Secret (required when authenticating with the official Bitwarden instances). Do not commit this information in Git unless you know what you're doing. Prefer using a Terraform `variable {}` in order to inject this value from the environment."
	descriptionServer         = "Bitwarden server URL (default: https://vault.bitwarden.com)."
	descriptionEmail          = "Login Email of the Vault."
	descriptionVaultPath      = "Alternative directory for storing the Vault locally."
)
