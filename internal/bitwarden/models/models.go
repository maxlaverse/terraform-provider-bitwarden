package models

import (
	"errors"
	"time"
)

var (
	ErrObjectNotFound      = errors.New("object not found")
	ErrAttachmentNotFound  = errors.New("attachment not found")
	ErrVaultLocked         = errors.New("vault is locked")
	ErrWrongMasterPassword = errors.New("invalid master password")
)

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
	ObjectTypeItem                ObjectType = "item"
	ObjectTypeAttachment          ObjectType = "attachment"
	ObjectTypeFolder              ObjectType = "folder"
	ObjectTypeOrgCollection       ObjectType = "org-collection"
	ObjectTypeOrganization        ObjectType = "organization"
	ObjectTypeList                ObjectType = "list"              // encapsulates collection list response
	ObjectTypeCollectionDetails   ObjectType = "collectionDetails" // collection listed in sync
	ObjectTypeCollection          ObjectType = "collection"        // used when refetching collections
	ObjectTypeProfile             ObjectType = "profile"
	ObjectTypeSync                ObjectType = "sync"
	ObjectTypeProfileOrganization ObjectType = "profileOrganization"   // organization under profile
	ObjectCipherDetails           ObjectType = "cipherDetails"         // when creating attachment data
	ObjectAttachmentFileUpload    ObjectType = "attachment-fileUpload" // when creating attachment data
)

const (
	DateLayout = "2006-01-02T15:04:05.000Z"
)

type Login struct {
	Username string     `json:"username,omitempty"`
	Password string     `json:"password,omitempty"`
	Totp     string     `json:"totp,omitempty"`
	URIs     []LoginURI `json:"uris,omitempty"`
}

type URIMatch int

const (
	URIMatchBaseDomain URIMatch = 0
	URIMatchHost       URIMatch = 1
	URIMatchStartWith  URIMatch = 2
	URIMatchExact      URIMatch = 3
	URIMatchRegExp     URIMatch = 4
	URIMatchNever      URIMatch = 5
)

func (u URIMatch) ToPointer() *URIMatch {
	return &u
}

type LoginURI struct {
	Match *URIMatch `json:"match,omitempty"`
	URI   string    `json:"uri,omitempty"`
}

type SecureNote struct {
	Type int `json:"type,omitempty"`
}

type Object struct {
	Attachments         []Attachment          `json:"attachments,omitempty"`
	Card                []byte                `json:"-"`
	CollectionIds       []string              `json:"collectionIds,omitempty"`
	CreationDate        *time.Time            `json:"creationDate,omitempty"`
	DeletedDate         *time.Time            `json:"deletedDate,omitempty"`
	Edit                bool                  `json:"edit,omitempty"`
	Favorite            bool                  `json:"favorite,omitempty"`
	Fields              []Field               `json:"fields,omitempty"`
	FolderID            string                `json:"folderId,omitempty"`
	Groups              []interface{}         `json:"groups"` // To be kept for the CLI when creating org-collections
	ID                  string                `json:"id,omitempty"`
	Identity            string                `json:"-"`
	Key                 string                `json:"key,omitempty"`
	Login               Login                 `json:"login,omitempty"`
	Name                string                `json:"name,omitempty"`
	Notes               string                `json:"notes,omitempty"`
	Object              ObjectType            `json:"object,omitempty"`
	OrganizationID      string                `json:"organizationId,omitempty"`
	OrganizationUseTotp bool                  `json:"organizationUseTotp,omitempty"`
	PasswordHistory     []PasswordHistoryItem `json:"passwordHistory,omitempty"`
	Reprompt            int                   `json:"reprompt,omitempty"`
	RevisionDate        *time.Time            `json:"revisionDate,omitempty"`
	SecureNote          SecureNote            `json:"secureNote,omitempty"`
	Type                ItemType              `json:"type,omitempty"`
	ViewPassword        bool                  `json:"viewPassword,omitempty"`
}

type PasswordHistoryItem struct {
	LastUsedDate *time.Time `json:"lastUsedDate,omitempty"`
	Password     string     `json:"password,omitempty"`
}

type Field struct {
	Name     string    `json:"name,omitempty"`
	Value    string    `json:"value,omitempty"`
	Type     FieldType `json:"type"`
	LinkedId *int      `json:"linkedId"`
}

type Attachment struct {
	ID       string     `json:"id,omitempty"`
	FileName string     `json:"fileName,omitempty"`
	Size     string     `json:"size,omitempty"`
	SizeName string     `json:"sizeName,omitempty"`
	Url      string     `json:"url,omitempty"`
	Key      string     `json:"key"`
	Object   ObjectType `json:"object"`
}
