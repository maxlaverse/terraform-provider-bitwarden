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
	AttachmentId   string                `json:"attachmentId"`
	CipherResponse models.Item           `json:"cipherResponse"`
	FileUploadType models.FileUploadType `json:"fileUploadType"`
	Object         models.ObjectType     `json:"object"`
	Url            string                `json:"url"`
}

type CollectionUser struct {
	HidePasswords bool   `json:"hidePasswords"`
	Id            string `json:"id"`
	ReadOnly      bool   `json:"readOnly"`
	Manage        bool   `json:"manage"`
}

type InviteUserRequest struct {
	Emails      []string `json:"emails"`
	Collections []string `json:"collections"`
	AccessAll   bool     `json:"accessAll"`
	Permissions struct {
		Response interface{} `json:"response"`
	} `json:"permissions"`
	Type                 models.OrgMemberRoleType `json:"type"`
	Groups               []string                 `json:"groups"`
	AccessSecretsManager bool                     `json:"accessSecretsManager"`
}

type ConfirmUserRequest struct {
	Key string `json:"key"`
}

type UserPublicKeyResponse struct {
	Object    models.ObjectType `json:"object"`
	PublicKey string            `json:"publicKey"`
	UserId    string            `json:"userId"`
}

type OrganizationUserList struct {
	Data   []OrganizationUserDetails `json:"data"`
	Object models.ObjectType         `json:"object"`
}

type OrganizationGroupList struct {
	Data   []OrganizationGroupDetails `json:"data"`
	Object models.ObjectType          `json:"object"`
}

type OrganizationGroupDetails struct {
	AccessAll      bool              `json:"accessAll"`
	Collections    []Collection      `json:"collections"`
	Email          string            `json:"email"`
	ExternalId     string            `json:"externalId"`
	Id             string            `json:"id"`
	Name           string            `json:"name"`
	Object         models.ObjectType `json:"object"`
	OrganizationId string            `json:"organizationId"`
}

type OrganizationUserDetails struct {
	AccessAll            bool              `json:"accessAll"`
	AccessSecretsManager bool              `json:"accessSecretsManager"`
	AvatarColor          string            `json:"avatarColor"`
	Collections          []Collection      `json:"collections"`
	Email                string            `json:"email"`
	ExternalId           string            `json:"externalId"`
	Groups               []string          `json:"groups"`
	HasMasterPassword    bool              `json:"hasMasterPassword"`
	Id                   string            `json:"id"`
	Name                 string            `json:"name"`
	Object               models.ObjectType `json:"object"`
	Permissions          struct {
		AccessEventLogs           bool `json:"accessEventLogs"`
		AccessImportExport        bool `json:"accessImportExport"`
		AccessReports             bool `json:"accessReports"`
		CreateNewCollections      bool `json:"createNewCollections"`
		DeleteAnyCollection       bool `json:"deleteAnyCollection"`
		DeleteAssignedCollections bool `json:"deleteAssignedCollections"`
		EditAnyCollection         bool `json:"editAnyCollection"`
		EditAssignedCollections   bool `json:"editAssignedCollections"`
		ManageGroups              bool `json:"manageGroups"`
		ManagePolicies            bool `json:"managePolicies"`
		ManageResetPassword       bool `json:"manageResetPassword"`
		ManageScim                bool `json:"manageScim"`
		ManageSso                 bool `json:"manageSso"`
		ManageUsers               bool `json:"manageUsers"`
	} `json:"permissions"`
	ResetPasswordEnrolled bool   `json:"resetPasswordEnrolled"`
	SsoBound              bool   `json:"ssoBound"`
	Status                int    `json:"status"`
	TwoFactorEnabled      bool   `json:"twoFactorEnabled"`
	Type                  int    `json:"type"`
	UserId                string `json:"userId"`
	UsesKeyConnector      bool   `json:"usesKeyConnector"`
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
	Assigned       bool              `json:"assigned"`
	ExternalId     string            `json:"externalId"`
	Groups         []string          `json:"groups"`
	HidePasswords  bool              `json:"hidePasswords"` // Missing in get collections
	Id             string            `json:"id"`
	Manage         bool              `json:"manage"`
	Name           string            `json:"name"`
	Object         models.ObjectType `json:"object"`
	OrganizationId string            `json:"organizationId"`
	ReadOnly       bool              `json:"readOnly"` // Missing in get collections
	Unmanaged      bool              `json:"unmanaged"`
	Users          []CollectionUser  `json:"users"`
}

type CreateCipherRequest struct {
	Cipher        models.Item `json:"cipher"`
	CollectionIds []string    `json:"collectionIds"`
}

type SyncResponse struct {
	Ciphers     []models.Item     `json:"ciphers"`
	Collections []Collection      `json:"collections"`
	Folders     []models.Folder   `json:"folders"`
	Object      models.ObjectType `json:"object"`
	Profile     Profile           `json:"profile"`
}

type ApiKey struct {
	ApiKey       string            `json:"apiKey"`
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

type MachineTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpireIn         int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	EncryptedPayload string `json:"encrypted_payload"`

	// To load profile ? Not present with access login
	PrivateKey string `json:"private_key,omitempty"`
	Key        string `json:"key,omitempty"`
}

type MachineTokenEncryptedPayload struct {
	EncryptionKey string `json:"encryptionKey"`
}

type Projects struct {
	Data              []models.Project `json:"data"`
	ContinuationToken *string          `json:"continuationToken"`
	Object            string           `json:"object"`
}

type SecretSummary struct {
	ID             string           `json:"id"`
	OrganizationID string           `json:"organizationId"`
	Key            string           `json:"key"`
	CreationDate   time.Time        `json:"creationDate"`
	RevisionDate   time.Time        `json:"revisionDate"`
	Projects       []models.Project `json:"projects"`
	Read           bool             `json:"read"`
	Write          bool             `json:"write"`
}

type Secret struct {
	SecretSummary
	Value  string `json:"value"`
	Note   string `json:"note"`
	Object string `json:"object"`
}

type SecretsWithProjectsList struct {
	Secrets  []SecretSummary  `json:"secrets"`
	Projects []models.Project `json:"projects"`
	Object   string           `json:"object"`
}

type CreateSecretRequest struct {
	Key                    string                  `json:"key"`
	Value                  string                  `json:"value"`
	Note                   string                  `json:"note"`
	ProjectIDs             []string                `json:"projectIds"`
	AccessPoliciesRequests *AccessPoliciesRequests `json:"accessPoliciesRequests,omitempty"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type AccessPoliciesRequests struct {
	UserAccessPolicyRequests           []interface{} `json:"userAccessPolicyRequests"`
	GroupAccessPolicyRequests          []interface{} `json:"groupAccessPolicyRequests"`
	ServiceAccountAccessPolicyRequests []interface{} `json:"serviceAccountAccessPolicyRequests"`
}

type CollectionAccessResponse struct {
	ContinuationToken string            `json:"continuationToken"`
	Data              []Collection      `json:"data"`
	Object            models.ObjectType `json:"object"`
}

type ErrorResponse struct {
	Message               string        `json:"message"`
	ValidationErrors      []interface{} `json:"validationErrors"`
	ExceptionMessage      string        `json:"exceptionMessage"`
	ExceptionStackTrace   interface{}   `json:"exceptionStackTrace"`
	InnerExceptionMessage string        `json:"innerExceptionMessage"`
	Object                string        `json:"object"`
}
