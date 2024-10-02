package webapi

import (
	"crypto/rsa"
	"time"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

type SignupRequest struct {
	Email              string         `json:"email"`
	Name               string         `json:"name"`
	MasterPasswordHash string         `json:"masterPasswordHash"`
	Key                string         `json:"key"`
	Kdf                models.KdfType `json:"kdf"`
	KdfIterations      int            `json:"kdfIterations"`
	KdfMemory          int            `json:"kdfMemory"`
	KdfParallelism     int            `json:"kdfParallelism"`
	Keys               KeyPair        `json:"keys"`
}

type KeyPair struct {
	PublicKey           string `json:"publicKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
}

type AttachmentRequestData struct {
	Key      string `json:"key"`
	FileName string `json:"fileName"`
	FileSize int    `json:"fileSize"`
}

type CreateObjectAttachmentResponse struct {
	AttachmentId   string            `json:"attachmentId"`
	CipherResponse models.Object     `json:"cipherResponse"`
	FileUploadType int               `json:"fileUploadType"`
	Object         models.ObjectType `json:"object"`
	Url            string            `json:"url"`
}

type OrganizationUser struct {
	Id            string `json:"id"`
	HidePasswords bool   `json:"hidePasswords"`
	ReadOnly      bool   `json:"readOnly"`
}
type OrganizationCreationRequest struct {
	Name       string             `json:"name"`
	Users      []OrganizationUser `json:"users"`
	Groups     []string           `json:"groups"`
	ExternalID string             `json:"externalId"`
}

type CreateOrganizationRequest struct {
	Name           string  `json:"name"`
	BillingEmail   string  `json:"billingEmail"`
	PlanType       int     `json:"planType"`
	CollectionName string  `json:"collectionName"`
	Key            string  `json:"key"`
	Keys           KeyPair `json:"keys"`
}

type CreateOrganizationResponse struct {
	Id string `json:"id"`
}
type PreloginResponse struct {
	Kdf            models.KdfType `json:"kdf"`
	KdfIterations  int            `json:"kdfIterations"`
	KdfMemory      int            `json:"kdfMemory"`
	KdfParallelism int            `json:"kdfParallelism"`
}
type TokenResponse struct {
	Kdf                 models.KdfType `json:"Kdf"`
	KdfIterations       int            `json:"KdfIterations"`
	KdfMemory           int            `json:"kdfMemory"`
	KdfParallelism      int            `json:"kdfParallelism"`
	Key                 string         `json:"Key"`
	PrivateKey          string         `json:"PrivateKey"`
	ResetMasterPassword bool           `json:"ResetMasterPassword"`
	AccessToken         string         `json:"access_token"`
	ExpireIn            int            `json:"expires_in"`
	RefreshToken        string         `json:"refresh_token"`
	Scope               string         `json:"scope"`
	TokenType           string         `json:"token_type"`
	UnofficialServer    bool           `json:"unofficialServer"`
	RSAPrivateKey       *rsa.PrivateKey
}

type PreloginRequest struct {
	Email string `json:"email"`
}

type RegistrationResponse struct {
	CaptchaBypassToken string            `json:"captchaBypassToken"`
	Object             models.ObjectType `json:"object"`
}

type CollectionResponse struct {
	Data   []CollectionResponseItem `json:"data"`
	Object models.ObjectType        `json:"object"`
}

type CollectionResponseItem struct {
	Id             string            `json:"id"`
	Name           string            `json:"name"`
	OrganizationId int               `json:"organization_id"`
	Object         models.ObjectType `json:"object"`
	ExternalId     string            `json:"external_id"`
}

type Collection struct {
	ExternalId     string            `json:"externalId,omitempty"`
	Id             string            `json:"id,omitempty"`
	HidePasswords  bool              `json:"hidePasswords,omitempty"` // Missing in get collections
	Name           string            `json:"name,omitempty"`
	Object         models.ObjectType `json:"object,omitempty"`
	OrganizationId string            `json:"organizationId,omitempty"`
	ReadOnly       bool              `json:"readOnly,omitempty"` // Missing in get collections
}

type CreateCipherRequest struct {
	Cipher        models.Object `json:"cipher"`
	CollectionIds []string      `json:"collectionIds"`
}

type SyncResponse struct {
	Ciphers     []models.Object   `json:"ciphers"`
	Collections []Collection      `json:"collections"`
	Folders     []Folder          `json:"folders"`
	Object      models.ObjectType `json:"object"`
	Profile     Profile           `json:"profile"`
}

type ApiKey struct {
	ApiKey       string            `json:"apiKey"`
	Object       models.ObjectType `json:"object"`
	RevisionDate *time.Time        `json:"revisionDate"`
}

type Folder struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Object       models.ObjectType `json:"object"`
	RevisionDate *time.Time        `json:"revisionDate"`
}

type Profile struct {
	Email         string            `json:"email"`
	Id            string            `json:"id"`
	Key           string            `json:"key"`
	Name          string            `json:"name"`
	Object        models.ObjectType `json:"object"`
	Organizations []Organization    `json:"organizations"`
	PrivateKey    string            `json:"privateKey"`
}

type Organization struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}
