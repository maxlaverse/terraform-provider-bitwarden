package embedded

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

type SecretsManager interface {
	CreateSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	DeleteSecret(ctx context.Context, secret models.Secret) error
	EditSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	GetProject(ctx context.Context, project models.Project) (*models.Project, error)
	GetSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	LoginWithAccessToken(ctx context.Context, accessToken string) error
}
type SecretsManagerOptions func(c bitwarden.SecretsManager)

func WithSecretsManagerHttpOptions(opts ...webapi.Options) SecretsManagerOptions {
	return func(c bitwarden.SecretsManager) {
		c.(*secretsManager).clientOpts = opts
	}
}

func NewSecretsManagerClient(serverURL, deviceIdentifier, providerVersion string, opts ...SecretsManagerOptions) SecretsManager {
	c := &secretsManager{
		serverURL: serverURL,
	}

	for _, o := range opts {
		o(c)
	}

	c.client = webapi.NewClient(serverURL, deviceIdentifier, providerVersion, c.clientOpts...)

	return c
}

type secretsManager struct {
	client     webapi.Client
	clientOpts []webapi.Options

	mainEncryptionKey  *symmetrickey.Key
	mainOrganizationId string
	serverURL          string
}

func (v *secretsManager) CreateSecret(ctx context.Context, secret models.Secret) (*models.Secret, error) {
	if v.mainEncryptionKey == nil {
		return nil, models.ErrLoggedOut
	}

	var resSecret *models.Secret

	encSecret, err := encryptSecret(secret, *v.mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting secret for creation: %w", err)
	}
	encSecret.OrganizationID = v.mainOrganizationId

	resEncSecret, err := v.client.CreateSecret(ctx, *encSecret)

	if err != nil {
		return nil, fmt.Errorf("error creating secret: %w", err)
	}

	resSecret, err = decryptSecret(*resEncSecret, *v.mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting secret after creation: %w", err)
	}

	return resSecret, nil
}

func (v *secretsManager) DeleteSecret(ctx context.Context, secret models.Secret) error {
	err := v.client.DeleteSecret(ctx, secret.ID)
	if err != nil {
		return fmt.Errorf("error deleting secret: %w", err)
	}

	return nil
}

func (v *secretsManager) EditSecret(ctx context.Context, secret models.Secret) (*models.Secret, error) {
	if v.mainEncryptionKey == nil {
		return nil, models.ErrLoggedOut
	}

	var resSecret *models.Secret

	encSecret, err := encryptSecret(secret, *v.mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting secret for edition: %w", err)
	}

	resEncSecret, err := v.client.EditSecret(ctx, *encSecret)

	if err != nil {
		return nil, fmt.Errorf("error editing secret: %w", err)
	}

	resSecret, err = decryptSecret(*resEncSecret, *v.mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting secret after edition: %w", err)
	}

	return resSecret, nil
}

func (v *secretsManager) GetProject(ctx context.Context, project models.Project) (*models.Project, error) {
	if v.mainEncryptionKey == nil {
		return nil, models.ErrLoggedOut
	}

	rawProject, err := v.client.GetProject(ctx, project.ID)
	if err != nil {
		if strings.Contains(err.Error(), "404!=200") {
			return nil, models.ErrObjectNotFound
		}
		return nil, fmt.Errorf("error getting project '%s': %w", project.ID, err)
	}

	decProject, err := decryptProject(*rawProject, *v.mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting project '%s': %w", project.ID, err)
	}
	decProject.OrganizationID = rawProject.OrganizationID
	return decProject, nil
}

func (v *secretsManager) GetSecret(ctx context.Context, secret models.Secret) (*models.Secret, error) {
	if v.mainEncryptionKey == nil {
		return nil, models.ErrLoggedOut
	}

	rawSecret, err := v.client.GetSecret(ctx, secret.ID)
	if err != nil {
		if strings.Contains(err.Error(), "404!=200") {
			return nil, models.ErrObjectNotFound
		}
		return nil, fmt.Errorf("error getting secret '%s': %w", secret.ID, err)
	}

	decSecret, err := decryptSecret(*rawSecret, *v.mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting secret '%s': %w", secret.ID, err)
	}
	decSecret.ProjectID = rawSecret.Projects[0].ID
	decSecret.OrganizationID = rawSecret.OrganizationID
	return decSecret, nil
}

func (v *secretsManager) LoginWithAccessToken(ctx context.Context, accessToken string) error {
	clientId, clientSecret, accessKeyEncryptionKey, err := parseAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("error parsing access token: %w", err)
	}

	tokenResp, err := v.client.LoginWithAccessToken(ctx, clientId, clientSecret)
	if err != nil {
		return fmt.Errorf("error login with access token: %w", err)
	}

	decryptedPayloadRaw, err := decryptStringAsBytes(tokenResp.EncryptedPayload, *accessKeyEncryptionKey)
	if err != nil {
		return fmt.Errorf("error decrypting encrypted payload: %w", err)
	}

	decryptedPayload := webapi.MachineTokenEncryptedPayload{}
	err = json.Unmarshal(decryptedPayloadRaw, &decryptedPayload)
	if err != nil {
		return fmt.Errorf("error unmarshalling encrypted payload: %w", err)
	}

	mainEncryptionKeyBytes, err := base64.StdEncoding.DecodeString(decryptedPayload.EncryptionKey)
	if err != nil {
		return fmt.Errorf("error decoding encryption key: %w", err)
	}

	mainEncryptionKey, err := symmetrickey.NewFromRawBytes(mainEncryptionKeyBytes)
	if err != nil {
		return fmt.Errorf("error loading encryption key: %w", err)
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenResp.AccessToken, &MachineAccountClaims{})
	if err != nil {
		return fmt.Errorf("error parsing token claims: %w", err)
	}

	if claims, ok := token.Claims.(*MachineAccountClaims); ok {
		v.mainOrganizationId = claims.Organization
	} else {
		return fmt.Errorf("invalid token: %w", err)
	}

	if len(v.mainOrganizationId) == 0 {
		return fmt.Errorf("organization ID not found in token claims")
	}
	v.mainEncryptionKey = mainEncryptionKey

	return nil
}

func decryptSecret(webapiSecret webapi.Secret, mainEncryptionKey symmetrickey.Key) (*models.Secret, error) {
	secretKey, err := decryptStringIfNotEmpty(webapiSecret.Key, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting secret name: %w", err)
	}

	secretNote, err := decryptStringIfNotEmpty(webapiSecret.Note, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting secret note: %w", err)
	}

	secretValue, err := decryptStringIfNotEmpty(webapiSecret.Value, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting secret value: %w", err)
	}

	projectId := ""
	if len(webapiSecret.Projects) > 0 {
		projectId = webapiSecret.Projects[0].ID
	}
	return &models.Secret{
		CreationDate:   webapiSecret.CreationDate,
		ID:             webapiSecret.ID,
		Key:            secretKey,
		Note:           secretNote,
		RevisionDate:   webapiSecret.RevisionDate,
		Value:          secretValue,
		ProjectID:      projectId,
		OrganizationID: webapiSecret.OrganizationID,
	}, nil
}

func decryptProject(webapiProject webapi.Project, mainEncryptionKey symmetrickey.Key) (*models.Project, error) {
	projectName, err := decryptStringIfNotEmpty(webapiProject.Name, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting project name: %w", err)
	}

	return &models.Project{
		CreationDate:   webapiProject.CreationDate,
		ID:             webapiProject.ID,
		Name:           projectName,
		OrganizationID: webapiProject.OrganizationID,
		RevisionDate:   webapiProject.RevisionDate,
	}, nil
}

func encryptSecret(secret models.Secret, mainEncryptionKey symmetrickey.Key) (*models.Secret, error) {
	secretKey, err := encryptAsStringIfNotEmpty(secret.Key, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypt secret name: %w", err)
	}

	secretNote, err := encryptAsStringIfNotEmpty(secret.Note, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypt secret note: %w", err)
	}

	secretValue, err := encryptAsStringIfNotEmpty(secret.Value, mainEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypt secret value: %w", err)
	}

	return &models.Secret{
		ID:             secret.ID,
		Key:            secretKey,
		Note:           secretNote,
		Value:          secretValue,
		ProjectID:      secret.ProjectID,
		OrganizationID: secret.OrganizationID,
	}, nil
}

func parseAccessToken(accessToken string) (string, string, *symmetrickey.Key, error) {
	accessTokenParts1 := strings.Split(accessToken, ":")
	if len(accessTokenParts1) != 2 {
		return "", "", nil, fmt.Errorf("invalid access token format (%d parts)", len(accessTokenParts1))
	}

	credentialsPart := accessTokenParts1[0]
	encryptionPart := accessTokenParts1[1]

	accessTokenParts := strings.Split(credentialsPart, ".")
	if len(accessTokenParts) != 3 {
		return "", "", nil, fmt.Errorf("invalid access token format (%d subparts)", len(accessTokenParts))
	}

	version := accessTokenParts[0]
	if version != "0" {
		return "", "", nil, fmt.Errorf("unsupported access token version: %s", version)
	}
	clientId := accessTokenParts[1]
	clientSecret := accessTokenParts[2]

	userEncryptionKey, err := keybuilder.DeriveFromAccessTokenEncryptionKey(encryptionPart)
	if err != nil {
		return "", "", nil, fmt.Errorf("error creating symmetric key: %w", err)
	}

	return clientId, clientSecret, userEncryptionKey, nil
}
