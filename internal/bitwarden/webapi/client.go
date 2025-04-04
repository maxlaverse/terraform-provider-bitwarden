package webapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

const (
	defaultRequestTimeout = 10 * time.Second
	maxConcurrentRequests = 4
	maxRetryAttempts      = 3
)

type Client interface {
	ClearSession()
	ConfirmOrganizationUser(ctx context.Context, orgID, orgUserID, key string) error
	CreateFolder(ctx context.Context, obj models.Folder) (*models.Folder, error)
	CreateItem(context.Context, models.Item) (*models.Item, error)
	CreateObjectAttachment(ctx context.Context, itemId string, data []byte, req AttachmentRequestData) (*CreateObjectAttachmentResponse, error)
	CreateObjectAttachmentData(ctx context.Context, itemId, attachmentId string, data []byte) error
	CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*CreateOrganizationResponse, error)
	CreateOrganizationCollection(ctx context.Context, orgId string, req Collection) (*Collection, error)
	CreateProject(ctx context.Context, project models.Project) (*models.Project, error)
	CreateSecret(ctx context.Context, secret models.Secret) (*Secret, error)
	DeleteFolder(ctx context.Context, objID string) error
	DeleteObject(ctx context.Context, objID string) error
	DeleteObjectAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteOrganizationCollection(ctx context.Context, orgID, collectionID string) error
	DeleteProject(ctx context.Context, projectId string) error
	DeleteSecret(ctx context.Context, secretId string) error
	EditFolder(ctx context.Context, obj models.Folder) (*models.Folder, error)
	EditItem(context.Context, models.Item) (*models.Item, error)
	EditItemCollections(ctx context.Context, objId string, collectionIds []string) (*models.Item, error)
	EditOrganizationCollection(ctx context.Context, orgId, objId string, obj Collection) (*Collection, error)
	EditProject(context.Context, models.Project) (*models.Project, error)
	EditSecret(ctx context.Context, secret models.Secret) (*Secret, error)
	GetAPIKey(ctx context.Context, username, password string, kdfConfig models.KdfConfiguration) (*ApiKey, error)
	GetOrganizationCollections(ctx context.Context, orgID string) ([]Collection, error)
	GetContentFromURL(ctx context.Context, url string) ([]byte, error)
	GetCipherAttachment(ctx context.Context, itemId, attachmentId string) (*models.Attachment, error)
	GetOrganizationUsers(ctx context.Context, orgId string) ([]OrganizationUserDetails, error)
	GetProfile(context.Context) (*Profile, error)
	GetProject(ctx context.Context, projectId string) (*models.Project, error)
	GetProjects(ctx context.Context, orgId string) ([]models.Project, error)
	GetSecret(ctx context.Context, secretId string) (*Secret, error)
	GetSecrets(ctx context.Context, orgId string) ([]SecretSummary, error)
	GetUserPublicKey(ctx context.Context, userId string) ([]byte, error)
	InviteUser(ctx context.Context, orgId string, user InviteUserRequest) error
	LoginWithAccessToken(ctx context.Context, clientId, clientSecret string) (*MachineTokenResponse, error)
	LoginWithAPIKey(ctx context.Context, clientId, clientSecret string) (*TokenResponse, error)
	LoginWithPassword(ctx context.Context, username, password string, kdfConfig models.KdfConfiguration) (*TokenResponse, error)
	PreLogin(context.Context, string) (*PreloginResponse, error)
	RegisterUser(ctx context.Context, req SignupRequest) error
	Sync(ctx context.Context) (*SyncResponse, error)
}

func NewClient(serverURL, deviceIdentifier, providerVersion string, opts ...Options) Client {
	c := &client{
		device:    DeviceInformation(deviceIdentifier, providerVersion),
		serverURL: strings.TrimSuffix(serverURL, "/"),
		httpClient: &http.Client{
			Transport: NewRetryRoundTripper(maxConcurrentRequests, maxRetryAttempts, defaultRequestTimeout),
		},
	}
	for _, o := range opts {
		o(c)
	}

	return c
}

type client struct {
	device             deviceInfoWithOfficialFallback
	httpClient         *http.Client
	serverURL          string
	sessionAccessToken string
}

func (c *client) ClearSession() {
	c.sessionAccessToken = ""
}

func (c *client) ConfirmOrganizationUser(ctx context.Context, orgID, orgUserId, key string) error {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations/%s/users/%s/confirm", c.serverURL, orgID, orgUserId), ConfirmUserRequest{Key: key})
	if err != nil {
		return fmt.Errorf("error preparing organization user confirmation request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) CreateFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/folders", c.serverURL), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing folder create request: %w", err)
	}

	return doRequest[models.Folder](ctx, c.httpClient, httpReq)
}

func (c *client) CreateItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	var err error
	var httpReq *http.Request
	if len(obj.CollectionIds) != 0 {
		cipherCreationRequest := CreateCipherRequest{
			Cipher:        obj,
			CollectionIds: obj.CollectionIds,
		}
		httpReq, err = c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/ciphers/create", c.serverURL), cipherCreationRequest)
	} else {
		httpReq, err = c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/ciphers", c.serverURL), obj)
	}
	if err != nil {
		return nil, fmt.Errorf("error preparing object create request: %w", err)
	}

	return doRequest[models.Item](ctx, c.httpClient, httpReq)
}

func (c *client) CreateObjectAttachment(ctx context.Context, itemId string, data []byte, req AttachmentRequestData) (*CreateObjectAttachmentResponse, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/ciphers/%s/attachment/v2", c.serverURL, itemId), req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshall attachment creation request: %w", err)
	}

	return doRequest[CreateObjectAttachmentResponse](ctx, c.httpClient, httpReq)
}

func (c *client) CreateObjectAttachmentData(ctx context.Context, itemId, attachmentId string, data []byte) error {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("data", "filename")
	if err != nil {
		return fmt.Errorf("error creating formfile: %w", err)
	}

	_, err = part.Write(data)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing writer: %w", err)
	}

	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/ciphers/%s/attachment/%s", c.serverURL, itemId, attachmentId), requestBody.Bytes())
	if err != nil {
		return fmt.Errorf("error preparing attachment create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations", c.serverURL), req)
	if err != nil {
		return nil, fmt.Errorf("error preparing organization creation request: %w", err)
	}

	return doRequest[CreateOrganizationResponse](ctx, c.httpClient, httpReq)
}

func (c *client) CreateOrganizationCollection(ctx context.Context, orgId string, req Collection) (*Collection, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations/%s/collections", c.serverURL, orgId), req)
	if err != nil {
		return nil, fmt.Errorf("error preparing organization collection creation request: %w", err)
	}

	return doRequest[Collection](ctx, c.httpClient, httpReq)
}

func (c *client) CreateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	projectCreationRequest := CreateProjectRequest{
		Name: project.Name,
	}
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations/%s/projects", c.serverURL, project.OrganizationID), projectCreationRequest)
	if err != nil {
		return nil, fmt.Errorf("error preparing secret creation request: %w", err)
	}

	return doRequest[models.Project](ctx, c.httpClient, httpReq)
}

func (c *client) CreateSecret(ctx context.Context, secret models.Secret) (*Secret, error) {
	cipherCreationRequest := CreateSecretRequest{
		Key:        secret.Key,
		Value:      secret.Value,
		Note:       secret.Note,
		ProjectIDs: []string{secret.ProjectID},
	}
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations/%s/secrets", c.serverURL, secret.OrganizationID), cipherCreationRequest)
	if err != nil {
		return nil, fmt.Errorf("error preparing secret creation request: %w", err)
	}

	return doRequest[Secret](ctx, c.httpClient, httpReq)
}

func (c *client) DeleteFolder(ctx context.Context, objID string) error {
	httpReq, err := c.prepareRequest(ctx, "DELETE", fmt.Sprintf("%s/api/folders/%s", c.serverURL, objID), nil)
	if err != nil {
		return fmt.Errorf("error preparing folder deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) DeleteObject(ctx context.Context, objID string) error {
	httpReq, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/ciphers/%s/delete", c.serverURL, objID), nil)
	if err != nil {
		return fmt.Errorf("error preparing object deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) DeleteObjectAttachment(ctx context.Context, itemId, attachmentId string) error {
	httpReq, err := c.prepareRequest(ctx, "DELETE", fmt.Sprintf("%s/api/ciphers/%s/attachment/%s", c.serverURL, itemId, attachmentId), nil)
	if err != nil {
		return fmt.Errorf("error preparing object attachment deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) DeleteOrganizationCollection(ctx context.Context, orgID, collectionID string) error {
	httpReq, err := c.prepareRequest(ctx, "DELETE", fmt.Sprintf("%s/api/organizations/%s/collections/%s", c.serverURL, orgID, collectionID), nil)
	if err != nil {
		return fmt.Errorf("error preparing organization collection deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) DeleteProject(ctx context.Context, projectId string) error {
	IDs := []string{projectId}
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/projects/delete", c.serverURL), IDs)
	if err != nil {
		return fmt.Errorf("error preparing project deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) DeleteSecret(ctx context.Context, secretId string) error {
	IDs := []string{secretId}
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/secrets/delete", c.serverURL), IDs)
	if err != nil {
		return fmt.Errorf("error preparing secret deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) GetCipherAttachment(ctx context.Context, itemId, attachmentId string) (*models.Attachment, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/ciphers/%s/attachment/%s", c.serverURL, itemId, attachmentId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing item attachment retrieval request: %w", err)
	}
	return doRequest[models.Attachment](ctx, c.httpClient, httpReq)
}

func (c *client) EditFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/folders/%s", c.serverURL, obj.ID), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing folder edition request: %w", err)
	}

	return doRequest[models.Folder](ctx, c.httpClient, req)
}

func (c *client) EditItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/ciphers/%s", c.serverURL, obj.ID), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing item edition request: %w", err)
	}

	return doRequest[models.Item](ctx, c.httpClient, req)
}

func (c *client) EditItemCollections(ctx context.Context, objId string, collectionIds []string) (*models.Item, error) {
	type CollectionChange struct {
		CollectionIDs []string `json:"collectionIds"`
	}

	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/ciphers/%s/collections_v2", c.serverURL, objId), CollectionChange{
		CollectionIDs: collectionIds,
	})
	if err != nil {
		return nil, fmt.Errorf("error preparing item collection edition request: %w", err)
	}

	return doRequest[models.Item](ctx, c.httpClient, req)
}

func (c *client) EditOrganizationCollection(ctx context.Context, orgId, objId string, obj Collection) (*Collection, error) {
	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/organizations/%s/collections/%s", c.serverURL, orgId, objId), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing collection edition request: %w", err)
	}

	return doRequest[Collection](ctx, c.httpClient, req)
}

func (c *client) EditProject(ctx context.Context, project models.Project) (*models.Project, error) {
	projectEditionRequest := CreateProjectRequest{
		Name: project.Name,
	}
	httpReq, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/projects/%s", c.serverURL, project.ID), projectEditionRequest)
	if err != nil {
		return nil, fmt.Errorf("error preparing project edition request: %w", err)
	}

	return doRequest[models.Project](ctx, c.httpClient, httpReq)
}

func (c *client) EditSecret(ctx context.Context, secret models.Secret) (*Secret, error) {
	cipherCreationRequest := CreateSecretRequest{
		Key:        secret.Key,
		Value:      secret.Value,
		Note:       secret.Note,
		ProjectIDs: []string{secret.ProjectID},
	}
	httpReq, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/secrets/%s", c.serverURL, secret.ID), cipherCreationRequest)
	if err != nil {
		return nil, fmt.Errorf("error preparing secret edition request: %w", err)
	}

	return doRequest[Secret](ctx, c.httpClient, httpReq)
}

func (c *client) GetAPIKey(ctx context.Context, username, password string, kdfConfig models.KdfConfiguration) (*ApiKey, error) {
	type ApiKeyRequest struct {
		MasterPasswordHash string `json:"masterPasswordHash"`
	}

	preloginKey, err := keybuilder.BuildPreloginKey(password, username, kdfConfig)
	if err != nil {
		return nil, fmt.Errorf("error building prelogin key: %w", err)
	}

	hashedPassword := crypto.HashPassword(password, *preloginKey, false)
	obj := ApiKeyRequest{MasterPasswordHash: hashedPassword}
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/accounts/api-key", c.serverURL), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing api key retrieval request: %w", err)
	}

	return doRequest[ApiKey](ctx, c.httpClient, httpReq)
}

func (c *client) GetContentFromURL(ctx context.Context, url string) ([]byte, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing raw URL retrieval request: %w", err)
	}

	resp, err := doRequest[[]byte](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return []byte(*resp), err
}

func (c *client) GetOrganizationCollections(ctx context.Context, orgID string) ([]Collection, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/organizations/%s/collections/details", c.serverURL, orgID), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing collection retrieval request: %w", err)
	}

	resp, err := doRequest[CollectionAccessResponse](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *client) GetProfile(ctx context.Context) (*Profile, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/accounts/profile", c.serverURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing config retrieval request: %w", err)
	}

	return doRequest[Profile](ctx, c.httpClient, httpReq)
}

func (c *client) GetProject(ctx context.Context, projectId string) (*models.Project, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/projects/%s", c.serverURL, projectId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing project retrieval request: %w", err)
	}

	return doRequest[models.Project](ctx, c.httpClient, httpReq)
}

func (c *client) GetProjects(ctx context.Context, orgId string) ([]models.Project, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/organizations/%s/projects", c.serverURL, orgId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing projects retrieval request: %w", err)
	}

	projects, err := doRequest[Projects](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	if projects.ContinuationToken != nil && len(*projects.ContinuationToken) > 0 {
		return nil, fmt.Errorf("pagination not supported")
	}
	return projects.Data, nil
}

func (c *client) GetSecret(ctx context.Context, secretId string) (*Secret, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/secrets/%s", c.serverURL, secretId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing secret retrieval request: %w", err)
	}

	return doRequest[Secret](ctx, c.httpClient, httpReq)
}

func (c *client) GetSecrets(ctx context.Context, orgId string) ([]SecretSummary, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/organizations/%s/secrets", c.serverURL, orgId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing secrets retrieval request: %w", err)
	}

	secrets, err := doRequest[SecretsWithProjectsList](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return secrets.Secrets, nil
}

func (c *client) GetUserPublicKey(ctx context.Context, userId string) ([]byte, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/users/%s/public-key", c.serverURL, userId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing user public key retrieval request: %w", err)
	}

	resp, err := doRequest[UserPublicKeyResponse](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return []byte(resp.PublicKey), nil
}

func (c *client) GetOrganizationUsers(ctx context.Context, orgId string) ([]OrganizationUserDetails, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/organizations/%s/users", c.serverURL, orgId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing organization user list retrieval request: %w", err)
	}

	resp, err := doRequest[OrganizationUserList](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *client) InviteUser(ctx context.Context, orgId string, inviteRequest InviteUserRequest) error {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations/%s/users/invite", c.serverURL, orgId), inviteRequest)
	if err != nil {
		return fmt.Errorf("error preparing user invitation request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) LoginWithAccessToken(ctx context.Context, clientId, clientSecret string) (*MachineTokenResponse, error) {
	form := url.Values{}
	form.Add("scope", "api.secrets")
	form.Add("client_id", clientId)
	form.Add("client_secret", clientSecret)
	form.Add("grant_type", "client_credentials")

	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/identity/connect/token", c.serverURL), form)
	if err != nil {
		return nil, fmt.Errorf("error preparing login with access token request: %w", err)
	}

	tokenResp, err := doRequest[MachineTokenResponse](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}

	c.sessionAccessToken = tokenResp.AccessToken

	return tokenResp, err
}

func (c *client) LoginWithPassword(ctx context.Context, username, password string, kdfConfig models.KdfConfiguration) (*TokenResponse, error) {
	preloginKey, err := keybuilder.BuildPreloginKey(password, username, kdfConfig)
	if err != nil {
		return nil, fmt.Errorf("error building prelogin key: %w", err)
	}

	hashedPassword := crypto.HashPassword(password, *preloginKey, false)

	form := url.Values{}
	form.Add("scope", "api offline_access")
	form.Add("client_id", "cli")
	form.Add("grant_type", "password")
	form.Add("username", username)
	form.Add("password", hashedPassword)

	// NOTE: The following fields are not documented on the Bitwarden API, but seem to be required
	//       in order to avoid a "No device information provided." error.
	form.Add("deviceType", c.device.official.deviceType)
	form.Add("deviceIdentifier", c.device.official.deviceIdentifier)
	form.Add("deviceName", c.device.official.deviceName)

	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/identity/connect/token", c.serverURL), form)
	if err != nil {
		return nil, fmt.Errorf("error preparing login with password request: %w", err)
	}

	// NOTE: There seem to be extra controls on the username/password authentication of the official
	//       Bitwarden server. Without a valid combination of headers, the server returns a 500
	//       after ~1 minute.
	httpReq.Header.Set("device-type", c.device.official.deviceType)
	httpReq.Header.Set("user-agent", c.device.official.userAgent)
	httpReq.Header.Set("auth-email", base64.RawURLEncoding.EncodeToString([]byte(username)))
	httpReq.Header.Set("bitwarden-client-name", "cli")

	tokenResp, err := doRequest[TokenResponse](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}

	encryptionKey, err := crypto.DecryptEncryptionKey(tokenResp.Key, *preloginKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting encryption key: %w", err)
	}

	privateKey, err := crypto.DecryptPrivateKey(tokenResp.PrivateKey, *encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private key: %w", err)
	}

	tokenResp.RSAPrivateKey = privateKey
	c.sessionAccessToken = tokenResp.AccessToken
	return tokenResp, nil
}

func (c *client) LoginWithAPIKey(ctx context.Context, clientId, clientSecret string) (*TokenResponse, error) {
	form := url.Values{}
	form.Add("scope", "api")
	form.Add("client_id", clientId)
	form.Add("client_secret", clientSecret)
	form.Add("grant_type", "client_credentials")

	// NOTE: The following fields are not documented on the Bitwarden API, but seem to be required
	//       in order to avoid a "No device information provided." error.
	form.Add("deviceType", c.device.deviceType)
	form.Add("deviceIdentifier", c.device.deviceIdentifier)
	form.Add("deviceName", c.device.deviceName)

	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/identity/connect/token", c.serverURL), form)
	if err != nil {
		return nil, fmt.Errorf("error preparing login with api key request: %w", err)
	}

	tokenResp, err := doRequest[TokenResponse](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	c.sessionAccessToken = tokenResp.AccessToken
	return tokenResp, err
}

func (c *client) PreLogin(ctx context.Context, username string) (*PreloginResponse, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/identity/accounts/prelogin", c.serverURL), PreloginRequest{Email: username})
	if err != nil {
		return nil, fmt.Errorf("error preparing prelogin request: %w", err)
	}

	return doRequest[PreloginResponse](ctx, c.httpClient, httpReq)
}

func (c *client) RegisterUser(ctx context.Context, signupRequest SignupRequest) error {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/accounts/register", c.serverURL), signupRequest)
	if err != nil {
		return fmt.Errorf("error preparing registration request: %w", err)
	}

	_, err = doRequest[RegistrationResponse](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) Sync(ctx context.Context) (*SyncResponse, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/sync?excludeDomains=true", c.serverURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing config retrieval request: %w", err)
	}

	return doRequest[SyncResponse](ctx, c.httpClient, httpReq)
}

func (c *client) prepareRequest(ctx context.Context, reqMethod, reqUrl string, reqBody interface{}) (*http.Request, error) {
	var httpReq *http.Request
	var err error

	if reqBody != nil {
		contentType := ""
		var bodyBytes []byte
		if v, ok := reqBody.(url.Values); ok {
			bodyBytes = []byte(v.Encode())
			contentType = "application/x-www-form-urlencoded; charset=utf-8"
		} else if v, ok := reqBody.([]byte); ok {
			bodyBytes = v
		} else {
			contentType = "application/json; charset=utf-8"
			bodyBytes, err = json.Marshal(reqBody)
			if err != nil {
				return nil, fmt.Errorf("unable to marshall request body: %w", err)
			}
		}
		httpReq, err = http.NewRequestWithContext(ctx, reqMethod, reqUrl, bytes.NewBuffer(bodyBytes))
		if httpReq != nil && len(contentType) > 0 {
			httpReq.Header.Add("Content-Type", contentType)
		}
	} else {
		httpReq, err = http.NewRequestWithContext(ctx, reqMethod, reqUrl, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error preparing http request: %w", err)
	}
	if len(c.sessionAccessToken) > 0 {
		httpReq.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.sessionAccessToken))
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("bitwarden-client-version", c.device.official.deviceVersion)

	if reqMethod == "GET" {
		httpReq.Header.Set("Cache-Control", "no-store")
		httpReq.Header.Set("Pragma", "no-cache")
	}

	return httpReq, nil
}

func doRequest[T any](ctx context.Context, httpClient *http.Client, httpReq *http.Request) (*T, error) {
	logRequest(ctx, httpReq)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error doing request to '%s': %w", httpReq.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from '%s %s': %w", httpReq.Method, httpReq.URL, err)
	}

	debugInfo := map[string]interface{}{"status_code": resp.StatusCode, "url": httpReq.URL.RequestURI(), "body": string(body), "headers": resp.Header}
	tflog.Trace(ctx, "Response from Bitwarden server", debugInfo)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response status code for '%s %s': %d!=200, body:%s", httpReq.Method, httpReq.URL, resp.StatusCode, string(body))
	}

	var res T
	if _, ok := any(&res).(*[]byte); ok {
		return any(&body).(*T), nil
	}

	if len(body) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Printf("Body to unmarshall: %s\n", string(body))
		return nil, fmt.Errorf("error unmarshalling response from '%s': %w", httpReq.URL, err)
	}

	return &res, nil
}

func logRequest(ctx context.Context, httpReq *http.Request) {
	debugInfo := map[string]interface{}{
		"url":     httpReq.URL.RequestURI(),
		"method":  httpReq.Method,
		"headers": httpReq.Header,
	}

	if httpReq.Body != nil {
		bodyCopy, newBody, err := readAndRestoreBody(httpReq.Body)
		if err != nil {
			tflog.Trace(ctx, "Unable to re-read request body", map[string]interface{}{"error": err})
		}
		httpReq.Body = newBody
		debugInfo["body"] = string(bodyCopy)
	}

	tflog.Trace(ctx, "Request to Bitwarden server ", debugInfo)
}

func readAndRestoreBody(rc io.ReadCloser) ([]byte, io.ReadCloser, error) {
	var buf bytes.Buffer

	tee := io.TeeReader(rc, &buf)

	body, err := io.ReadAll(tee)
	if err != nil {
		return nil, nil, err
	}
	return body, io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
