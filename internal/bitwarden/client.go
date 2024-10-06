package bitwarden

import (
	"context"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

const (
	DefaultBitwardenServerURL = "https://vault.bitwarden.com"
)

type PasswordManager interface {
	CreateAttachment(ctx context.Context, itemId, filePath string) (*models.Object, error)
	CreateObject(context.Context, models.Object) (*models.Object, error)
	DeleteAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteObject(context.Context, models.Object) error
	EditObject(context.Context, models.Object) (*models.Object, error)
	GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error)
	GetObject(context.Context, models.Object) (*models.Object, error)
	ListObjects(ctx context.Context, objType models.ObjectType, options ...ListObjectsOption) ([]models.Object, error)
	LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error
	LoginWithPassword(ctx context.Context, username, password string) error
	Sync(context.Context) error
}

type SecretsManager interface {
	CreateSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	DeleteSecret(ctx context.Context, secret models.Secret) error
	EditSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	GetProject(ctx context.Context, project models.Project) (*models.Project, error)
	GetSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	LoginWithAccessToken(ctx context.Context, accessKey string) error
}
