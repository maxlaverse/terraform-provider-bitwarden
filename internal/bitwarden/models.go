package bitwarden

import "time"

type ItemType int

const (
	ItemTypeLogin      ItemType = 1
	ItemTypeSecureNote ItemType = 2
)

type FieldType int

const (
	FieldTypeText    FieldType = 0
	FieldTypeHidden  FieldType = 1
	FieldTypeBoolean FieldType = 2
	FieldTypeLinked  FieldType = 3
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
}

type SecureNote struct {
	Type int `json:"type,omitempty"`
}

type Object struct {
	CollectionIds  []string   `json:"collectionIds,omitempty"`
	ID             string     `json:"id,omitempty"`
	ExternalID     string     `json:"externalId,omitempty"`
	FolderID       string     `json:"folderId,omitempty"`
	Login          Login      `json:"login,omitempty"`
	Name           string     `json:"name,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	Object         ObjectType `json:"object,omitempty"`
	OrganizationID string     `json:"organizationId,omitempty"`
	SecureNote     SecureNote `json:"secureNote,omitempty"`
	Type           ItemType   `json:"type,omitempty"`
	Fields         []Field    `json:"fields,omitempty"`
	Reprompt       int        `json:"reprompt,omitempty"`
	Favorite       bool       `json:"favorite,omitempty"`
	RevisionDate   *time.Time `json:"revisionDate,omitempty"`
}

const (
	RevisionDateLayout = "2006-01-02T15:04:05.000Z"
)

type Field struct {
	Name  string    `json:"name,omitempty"`
	Value string    `json:"value,omitempty"`
	Type  FieldType `json:"type,omitempty"`
}
