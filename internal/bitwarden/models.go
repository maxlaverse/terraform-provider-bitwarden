package bitwarden

import "time"

type ItemType int

const (
	ItemTypeLogin      ItemType = 1
	ItemTypeSecureNote ItemType = 2
)

type ObjectType string

const (
	ObjectTypeItem   ObjectType = "item"
	ObjectTypeFolder ObjectType = "folder"
)

type VaultStatus string

const (
	StatusAuthenticated   VaultStatus = "authenticated"
	StatusLocked          VaultStatus = "locked"
	StatusUnauthenticated VaultStatus = "unauthenticated"
)

type Status struct {
	ServerURL string      `json:"serverURL,omitempty"`
	LastSync  time.Time   `json:"lastSync,omitempty"`
	UserEmail string      `json:"userEmail,omitempty"`
	UserID    string      `json:"userID,omitempty"`
	Status    VaultStatus `json:"status,omitempty"`
}

type Login struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Totp     string `json:"totp,omitempty"`
	/*
	* Fields that are not supported yet
	 */
	// PasswordRevisionDate string `json:"passwordRevisionDate,omitempty"`
}

/*
* Item type that are not supported yet
 */
// type Card struct {
// 	CardholderName string `json:"cardholderName,omitempty"`
// 	Brand          string `json:"brand,omitempty"`
// 	Number         string `json:"number,omitempty"`
// 	ExpMonth       string `json:"expMonth,omitempty"`
// 	ExpYear        string `json:"expYear,omitempty"`
// 	Code           string `json:"code,omitempty"`
// }

type SecureNote struct {
	Type int `json:"type,omitempty"`
}

type Object struct {
	Object     ObjectType `json:"object,omitempty"`
	ID         string     `json:"id,omitempty"`
	FolderID   string     `json:"folderId,omitempty"`
	Type       ItemType   `json:"type,omitempty"`
	Name       string     `json:"name,omitempty"`
	Notes      string     `json:"notes,omitempty"`
	Login      Login      `json:"login,omitempty"`
	SecureNote SecureNote `json:"secureNote,omitempty"`

	/*
	* Fields that are not supported yet
	 */
	// Reprompt       int       `json:"reprompt,omitempty"`
	// Favorite       bool      `json:"favorite,omitempty"`
	// Card           Card      `json:"card,omitempty"`
	// CollectionIds  []string  `json:"collectionIds,omitempty"`
	// RevisionDate   time.Time `json:"revisionDate,omitempty"`
	OrganizationID string `json:"organizationId,omitempty"`
}
