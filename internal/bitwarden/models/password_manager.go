package models

import (
	"errors"
	"time"
)

var (
	ErrObjectNotFound              = errors.New("object not found")
	ErrAttachmentNotFound          = errors.New("attachment not found")
	ErrVaultLocked                 = errors.New("vault is locked")
	ErrAlreadyLoggedIn             = errors.New("you are already logged in")
	ErrWrongMasterPassword         = errors.New("invalid master password")
	ErrLoggedOut                   = errors.New("please login first")
	ErrItemTypeMismatch            = errors.New("returned object type does not match requested object type")
	ErrTooManyObjectsFound         = errors.New("too many objects found")
	ErrNoObjectFoundMatchingFilter = errors.New("no object found matching the filter")
)

type ItemType int

const (
	ItemTypeLogin      ItemType = 1
	ItemTypeSecureNote ItemType = 2
)

type KdfType int

const (
	KdfTypePBKDF2_SHA256 KdfType = 0
	KdfTypeArgon2        KdfType = 1
)

type KdfConfiguration struct {
	KdfIterations  int     `json:"kdfIterations,omitempty"`
	KdfMemory      int     `json:"kdfMemory,omitempty"`
	KdfParallelism int     `json:"kdfParallelism,omitempty"`
	KdfType        KdfType `json:"kdfType,omitempty"`
}

type FieldType int

const (
	FieldTypeText    FieldType = 0
	FieldTypeHidden  FieldType = 1
	FieldTypeBoolean FieldType = 2
	FieldTypeLinked  FieldType = 3
)

type OrgMemberRoleType int

const (
	// According to UI: Manage all aspects of your organization, including billing and subscriptions
	OrgMemberRoleTypeOwner OrgMemberRoleType = 0

	// According to UI: Manage organization access, all collections, members, reporting, and security settings
	OrgMemberRoleTypeAdmin OrgMemberRoleType = 1

	// According to UI: Access and add items to assigned collections
	OrgMemberRoleTypeUser OrgMemberRoleType = 2

	// According to UI: Create, delete, and manage access in assigned collections
	OrgMemberRoleTypeManager OrgMemberRoleType = 3
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
	ObjectTypeOrgMember           ObjectType = "org-member"
	ObjectTypeProfile             ObjectType = "profile"
	ObjectTypeSync                ObjectType = "sync"
	ObjectTypeProfileOrganization ObjectType = "profileOrganization"   // organization under profile
	ObjectCipherDetails           ObjectType = "cipherDetails"         // when creating attachment data
	ObjectAttachmentFileUpload    ObjectType = "attachment-fileUpload" // when creating attachment data
	ObjectApiKey                  ObjectType = "api-key"
	ObjectProject                 ObjectType = "project"
	ObjectSecret                  ObjectType = "secret"
	ObjectUserKey                 ObjectType = "userKey"
)

type FileUploadType int

const (
	FileUploadTypeDirect FileUploadType = 0
	FileUploadTypeAzure  FileUploadType = 1
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

type ApiKey struct {
	ClientID     string
	ClientSecret string
}

type Item struct {
	Attachments         []Attachment          `json:"attachments,omitempty"`
	CollectionIds       []string              `json:"collectionIds,omitempty"`
	CreationDate        *time.Time            `json:"creationDate,omitempty"`
	DeletedDate         *time.Time            `json:"deletedDate,omitempty"`
	Edit                bool                  `json:"edit,omitempty"`
	Favorite            bool                  `json:"favorite,omitempty"`
	Fields              []Field               `json:"fields,omitempty"`
	FolderID            string                `json:"folderId,omitempty"`
	ID                  string                `json:"id,omitempty"`
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

type Folder struct {
	ID           string     `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Object       ObjectType `json:"object,omitempty"`
	RevisionDate *time.Time `json:"revisionDate,omitempty"`
}

type Organization struct {
	ID     string     `json:"id,omitempty"`
	Name   string     `json:"name,omitempty"`
	Object ObjectType `json:"object,omitempty"`
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

type OrgCollectionMember struct {
	HidePasswords bool   `json:"hidePasswords"`
	Id            string `json:"id"`
	Manage        bool   `json:"manage"`
	ReadOnly      bool   `json:"readOnly"`
}

type OrgMember struct {
	OrganizationId string
	ID             string
	Email          string
	Name           string
	UserId         string
}

type OrgCollection struct {
	ID             string                `json:"id,omitempty"`
	Name           string                `json:"name,omitempty"`
	Object         ObjectType            `json:"object,omitempty"`
	OrganizationID string                `json:"organizationId"`
	Users          []OrgCollectionMember `json:"users"`
	Groups         []interface{}         `json:"groups"` // Required but not used when creating collections using the CLI
	Manage         bool                  `json:"-"`
}

type Group struct {
	AccessAll      bool                  `json:"accessAll"`
	Collections    []OrgCollectionMember `json:"collections"`
	ID             string                `json:"id,omitempty"`
	Name           string                `json:"name,omitempty"`
	OrganizationID string                `json:"organizationId"`
	Users          []OrgCollectionMember `json:"users"`
}
