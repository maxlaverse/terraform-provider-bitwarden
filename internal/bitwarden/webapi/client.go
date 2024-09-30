package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

const (
	deviceName = "Bitwarden Terraform Provider"
)

type Client interface {
	CreateFolder(ctx context.Context, obj Folder) (*Folder, error)
	CreateObject(context.Context, models.Object) (*models.Object, error)
	CreateObjectAttachment(ctx context.Context, itemId string, data []byte, req AttachmentRequestData) (*CreateObjectAttachmentResponse, error)
	CreateObjectAttachmentData(ctx context.Context, itemId, attachmentId string, data []byte) error
	CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*CreateOrganizationResponse, error)
	CreateOrgCollection(ctx context.Context, orgId string, req OrganizationCreationRequest) (*Collection, error)
	DeleteFolder(ctx context.Context, objID string) error
	DeleteObject(ctx context.Context, objID string) error
	DeleteObjectAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteOrgCollection(ctx context.Context, orgID, collectionID string) error
	EditFolder(ctx context.Context, obj Folder) (*Folder, error)
	EditObject(context.Context, models.Object) (*models.Object, error)
	EditOrgCollection(ctx context.Context, orgId, objId string, obj OrganizationCreationRequest) (*Collection, error)
	GetCollections(ctx context.Context, orgID string) ([]CollectionResponseItem, error)
	GetContentFromURL(ctx context.Context, url string) ([]byte, error)
	GetObjectAttachment(ctx context.Context, itemId, attachmentId string) (*models.Attachment, error)
	LoginWithAPIKey(ctx context.Context, clientId, clientSecret string) (*TokenResponse, error)
	LoginWithPassword(ctx context.Context, username, password string, kdfIterations int) (*TokenResponse, error)
	PreLogin(context.Context, string) (*PreloginResponse, error)
	Profile(context.Context) (*Profile, error)
	RegisterUser(ctx context.Context, name, username, password string, kdfIterations int) error
	Sync(ctx context.Context) (*SyncResponse, error)
}

type Options func(c Client)

func DisableRetries() Options {
	return func(c Client) {
		c.(*client).httpClient.RetryMax = 0
	}
}

func WithCustomClient(httpClient http.Client) Options {
	return func(c Client) {
		c.(*client).httpClient.HTTPClient = &httpClient
	}
}

func WithDeviceIdentifier(deviceIdentifier string) Options {
	return func(c Client) {
		c.(*client).deviceIdentifier = deviceIdentifier
	}
}

func NewClient(serverURL string, opts ...Options) Client {
	c := &client{
		deviceName: deviceName,
		deviceType: "10",
		serverURL:  strings.TrimSuffix(serverURL, "/"),
		httpClient: retryablehttp.NewClient(),
	}
	for _, o := range opts {
		o(c)
	}
	c.httpClient.Logger = nil

	return c
}

type client struct {
	deviceIdentifier   string
	deviceName         string
	deviceType         string
	httpClient         *retryablehttp.Client
	serverURL          string
	sessionAccessToken string
}

func (c *client) CreateFolder(ctx context.Context, obj Folder) (*Folder, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/folders", c.serverURL), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing folder create request: %w", err)
	}

	return doRequest[Folder](ctx, c.httpClient, httpReq)
}

func (c *client) CreateObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	var err error
	var httpReq *retryablehttp.Request
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

	return doRequest[models.Object](ctx, c.httpClient, httpReq)
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

func (c *client) CreateOrgCollection(ctx context.Context, orgId string, req OrganizationCreationRequest) (*Collection, error) {
	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/api/organizations/%s/collections", c.serverURL, orgId), req)
	if err != nil {
		return nil, fmt.Errorf("error preparing organization collection creation request: %w", err)
	}

	return doRequest[Collection](ctx, c.httpClient, httpReq)
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

func (c *client) DeleteOrgCollection(ctx context.Context, orgID, collectionID string) error {
	httpReq, err := c.prepareRequest(ctx, "DELETE", fmt.Sprintf("%s/api/organizations/%s/collections/%s", c.serverURL, orgID, collectionID), nil)
	if err != nil {
		return fmt.Errorf("error preparing organization collection deletion request: %w", err)
	}

	_, err = doRequest[[]byte](ctx, c.httpClient, httpReq)
	return err
}

func (c *client) GetObjectAttachment(ctx context.Context, itemId, attachmentId string) (*models.Attachment, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/ciphers/%s/attachment/%s", c.serverURL, itemId, attachmentId), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing object attachment retrieval request: %w", err)
	}
	return doRequest[models.Attachment](ctx, c.httpClient, httpReq)
}

func (c *client) EditFolder(ctx context.Context, obj Folder) (*Folder, error) {
	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/folders/%s", c.serverURL, obj.Id), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing folder edition request: %w", err)
	}

	return doRequest[Folder](ctx, c.httpClient, req)
}

func (c *client) EditObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/ciphers/%s", c.serverURL, obj.ID), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing object edition request: %w", err)
	}

	return doRequest[models.Object](ctx, c.httpClient, req)
}

func (c *client) EditOrgCollection(ctx context.Context, orgId, objId string, obj OrganizationCreationRequest) (*Collection, error) {
	req, err := c.prepareRequest(ctx, "PUT", fmt.Sprintf("%s/api/organizations/%s/collections/%s", c.serverURL, orgId, objId), obj)
	if err != nil {
		return nil, fmt.Errorf("error preparing collection edition request: %w", err)
	}

	return doRequest[Collection](ctx, c.httpClient, req)
}

func (c *client) GetContentFromURL(ctx context.Context, url string) ([]byte, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing raw URL retrieval request: %w", err)
	}

	resp, err := doRequest[[]byte](ctx, c.httpClient, httpReq)
	return []byte(*resp), err
}

func (c *client) GetCollections(ctx context.Context, orgID string) ([]CollectionResponseItem, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/organizations/%s/collections", c.serverURL, orgID), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing collection retrieval request: %w", err)
	}

	resp, err := doRequest[CollectionResponse](ctx, c.httpClient, httpReq)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *client) LoginWithPassword(ctx context.Context, username, password string, kdfIterations int) (*TokenResponse, error) {
	preloginKey, err := keybuilder.BuildPreloginKey(password, username, kdfIterations)
	if err != nil {
		return nil, fmt.Errorf("error building prelogin key: %w", err)
	}

	hashedPassword := crypto.HashPassword(password, *preloginKey, false)

	form := url.Values{}
	form.Add("scope", "api offline_access")
	form.Add("client_id", "web")
	form.Add("grant_type", "password")
	form.Add("username", username)
	form.Add("password", hashedPassword)
	form.Add("device_type", c.deviceType)
	form.Add("device_identifier", c.deviceIdentifier)
	form.Add("device_name", c.deviceName)

	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/identity/connect/token", c.serverURL), form)
	if err != nil {
		return nil, fmt.Errorf("error preparing login with password request: %w", err)
	}

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
	form.Add("device_type", c.deviceType)
	form.Add("device_identifier", c.deviceIdentifier)
	form.Add("device_name", c.deviceName)

	httpReq, err := c.prepareRequest(ctx, "POST", fmt.Sprintf("%s/identity/connect/token", c.serverURL), form)
	if err != nil {
		return nil, fmt.Errorf("error preparing login with api key request: %w", err)
	}

	tokenResp, err := doRequest[TokenResponse](ctx, c.httpClient, httpReq)
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

func (c *client) Profile(ctx context.Context) (*Profile, error) {
	httpReq, err := c.prepareRequest(ctx, "GET", fmt.Sprintf("%s/api/accounts/profile", c.serverURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing config retrieval request: %w", err)
	}

	return doRequest[Profile](ctx, c.httpClient, httpReq)
}

func (c *client) RegisterUser(ctx context.Context, name, username, password string, kdfIterations int) error {
	preloginKey, err := keybuilder.BuildPreloginKey(password, username, kdfIterations)
	if err != nil {
		return fmt.Errorf("error building prelogin key: %w", err)
	}

	hashedPassword := crypto.HashPassword(password, *preloginKey, false)

	encryptionKey, encryptedEncryptionKey, err := keybuilder.GenerateEncryptionKey(*preloginKey)
	if err != nil {
		return fmt.Errorf("error generating encryption key: %w", err)
	}

	publicKey, encryptedPrivateKey, err := keybuilder.GenerateKeyPair(*encryptionKey)
	if err != nil {
		return fmt.Errorf("error generating key pair: %w", err)
	}

	signupRequest := SignupRequest{
		Email:              username,
		Name:               name,
		MasterPasswordHash: hashedPassword,
		Key:                encryptedEncryptionKey,
		KdfIterations:      kdfIterations,
		Keys: KeyPair{
			PublicKey:           publicKey,
			EncryptedPrivateKey: encryptedPrivateKey,
		},
	}

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

func (c *client) prepareRequest(ctx context.Context, reqMethod, reqUrl string, reqBody interface{}) (*retryablehttp.Request, error) {
	var httpReq *retryablehttp.Request
	var err error
	contentType := ""
	if reqBody != nil {
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
		httpReq, err = retryablehttp.NewRequestWithContext(ctx, reqMethod, reqUrl, bytes.NewBuffer(bodyBytes))
		tflog.Trace(ctx, "Request to Bitwarden server", map[string]interface{}{"url": reqUrl, "method": reqMethod, "body": string(bodyBytes)})
	} else {
		httpReq, err = retryablehttp.NewRequestWithContext(ctx, reqMethod, reqUrl, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error preparing http request: %w", err)
	}
	httpReq.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.sessionAccessToken))
	if len(contentType) > 0 {
		httpReq.Header.Add("Content-Type", contentType)
	}

	return httpReq, nil
}

func doRequest[T any](ctx context.Context, httpClient *retryablehttp.Client, httpReq *retryablehttp.Request) (*T, error) {
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error doing request to '%s': %w", httpReq.URL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from to '%s': %w", httpReq.URL, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response status code for '%s': %d!=200, body:%s", httpReq.URL, resp.StatusCode, string(body))
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
	tflog.Trace(ctx, "Response from Bitwarden server", map[string]interface{}{"url": httpReq.URL, "body": string(body)})

	return &res, nil
}
