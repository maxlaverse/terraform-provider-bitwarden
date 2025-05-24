package bitwarden

import (
	"context"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

const (
	DefaultBitwardenServerURL = "https://vault.bitwarden.com"
)

type PasswordManager interface {
	CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Attachment, error)
	CreateAttachmentFromFile(ctx context.Context, itemId, filePath string) (*models.Attachment, error)
	CreateFolder(context.Context, models.Folder) (*models.Folder, error)
	CreateGroup(context.Context, models.Group) (*models.Group, error)
	CreateItem(context.Context, models.Item) (*models.Item, error)
	CreateOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	DeleteAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteFolder(context.Context, models.Folder) error
	DeleteGroup(context.Context, models.Group) error
	DeleteItem(context.Context, models.Item) error
	DeleteOrganizationCollection(context.Context, models.OrgCollection) error
	EditFolder(context.Context, models.Folder) (*models.Folder, error)
	EditGroup(context.Context, models.Group) (*models.Group, error)
	EditItem(context.Context, models.Item) (*models.Item, error)
	EditOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	FindFolder(ctx context.Context, options ...ListObjectsOption) (*models.Folder, error)
	FindItem(ctx context.Context, options ...ListObjectsOption) (*models.Item, error)
	FindOrganization(ctx context.Context, options ...ListObjectsOption) (*models.Organization, error)
	FindOrganizationMember(ctx context.Context, options ...ListObjectsOption) (*models.OrgMember, error)
	FindOrganizationCollection(ctx context.Context, options ...ListObjectsOption) (*models.OrgCollection, error)
	GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error)
	GetFolder(context.Context, models.Folder) (*models.Folder, error)
	GetGroup(context.Context, models.Group) (*models.Group, error)
	GetItem(context.Context, models.Item) (*models.Item, error)
	GetOrganization(context.Context, models.Organization) (*models.Organization, error)
	GetOrganizationMember(context.Context, models.OrgMember) (*models.OrgMember, error)
	GetOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error
	LoginWithPassword(ctx context.Context, username, password string) error
	Sync(context.Context) error
}

type SecretsManager interface {
	CreateProject(ctx context.Context, project models.Project) (*models.Project, error)
	CreateSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	DeleteProject(ctx context.Context, project models.Project) error
	DeleteSecret(ctx context.Context, secret models.Secret) error
	EditProject(ctx context.Context, project models.Project) (*models.Project, error)
	EditSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	GetProject(ctx context.Context, project models.Project) (*models.Project, error)
	GetSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	GetSecretByKey(ctx context.Context, secretKey string) (*models.Secret, error)
	LoginWithAccessToken(ctx context.Context, accessKey string) error
}
