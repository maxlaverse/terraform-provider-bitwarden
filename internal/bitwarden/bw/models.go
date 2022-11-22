package bw

import (
	"strings"
	"time"
)

type ItemType int

const (
	ItemTypeLogin      ItemType = 1
	ItemTypeSecureNote ItemType = 2
)

const (
	DefaultBitwardenServerURL = "https://vault.bitwarden.com"
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
	StatusLocked          VaultStatus = "locked"
	StatusUnauthenticated VaultStatus = "unauthenticated"
	StatusUnlocked        VaultStatus = "unlocked"
)

type Status struct {
	ServerURL string      `json:"serverURL,omitempty"`
	LastSync  time.Time   `json:"lastSync,omitempty"`
	UserEmail string      `json:"userEmail,omitempty"`
	UserID    string      `json:"userID,omitempty"`
	Status    VaultStatus `json:"status,omitempty"`
}

func (s *Status) VaultFromServer(serverUrl string) bool {
	providerServerUrl := trimSlashSuffix(serverUrl)
	vaultServerUrl := trimSlashSuffix(s.ServerURL)
	return vaultServerUrl == providerServerUrl || len(vaultServerUrl) == 0 && providerServerUrl == DefaultBitwardenServerURL
}

func (s *Status) VaultOfUser(email string) bool {
	return s.UserEmail == email
}

func (s *Status) FreshDataFile() bool {
	return len(s.UserEmail) == 0 && len(s.ServerURL) == 0 && s.Status == StatusUnauthenticated
}

func trimSlashSuffix(serverUrl string) string {
	return strings.TrimSuffix(serverUrl, "/")
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
