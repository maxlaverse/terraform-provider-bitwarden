package webapi

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
)

/*
* This is a client to interact with the Vaultwarden or eventually
* Bitwarden compatible API. It's only meant to be used for test
* purposes and is to be considered as an insecure implementation
 */

type Client interface {
	CreateOrganization(name, label, billingEmail string) (string, error)
	GetCollections(orgID string) (string, error)
	Login(username, password string, kdfIterations int) error
	RegisterUser(name, username, password string, kdfIterations int) error
}

type session struct {
	accessToken string
	privateKey  *rsa.PrivateKey
}

func NewClient(serverURL string) Client {
	return &client{
		deviceIdentifier: "5d90b470-5d1d-452d-935c-b730c177a8d6",
		deviceName:       "firefox",
		deviceType:       "10",
		serverURL:        serverURL,
	}
}

type client struct {
	deviceIdentifier string
	deviceName       string
	deviceType       string
	serverURL        string
	session          session
}

func (c *client) RegisterUser(name, username, password string, kdfIterations int) error {
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

	signupRequestBytes, err := json.Marshal(signupRequest)
	if err != nil {
		return fmt.Errorf("unable to marshall user registration request: %w", err)
	}

	resp, err := http.Post(c.signupURL(), "application/json", bytes.NewBuffer(signupRequestBytes))
	if err != nil {
		return fmt.Errorf("error calling user registration: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading registration response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code in registration response: %d!=200, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *client) Login(username, password string, kdfIterations int) error {
	preloginKey, err := keybuilder.BuildPreloginKey(password, username, kdfIterations)
	if err != nil {
		return fmt.Errorf("error building prelogin key: %w", err)
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

	req, err := http.NewRequest("POST", c.loginURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("error preparing login request: %w", err)
	}

	req.Header.Add("auth-email", base64.StdEncoding.EncodeToString([]byte(username)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Add("device-type", c.deviceType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling user login: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error during login call body reading: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code for login call: %d!=200, body:%s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return fmt.Errorf("error unmarshalling login response: %w", err)
	}

	encryptionKey, err := crypto.DecryptEncryptionKey(tokenResp.Key, *preloginKey)
	if err != nil {
		return fmt.Errorf("error decrypting encryption key: %w", err)
	}

	privateKey, err := crypto.DecryptPrivateKey(tokenResp.PrivateKey, *encryptionKey)
	if err != nil {
		return fmt.Errorf("error decrypting private key: %w", err)
	}

	c.session.accessToken = tokenResp.AccessToken
	c.session.privateKey = privateKey
	return nil
}

func (c *client) CreateOrganization(organizationName string, label string, billingEmail string) (string, error) {
	encryptedShareKey, shareKey, err := keybuilder.GenerateShareKey(&c.session.privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("error generating share key: %w", err)
	}

	collectionName, err := crypto.Encrypt([]byte(label), *shareKey)
	if err != nil {
		return "", fmt.Errorf("error encryption collection label: %w", err)
	}

	publicKey, encryptedPrivateKey, err := keybuilder.GenerateKeyPair(*shareKey)
	if err != nil {
		return "", fmt.Errorf("error generating key pair: %w", err)
	}

	orgCreationRequest := &CreateOrganizationRequest{
		Name:           organizationName,
		BillingEmail:   billingEmail,
		CollectionName: collectionName,
		Key:            encryptedShareKey,
		Keys: KeyPair{
			PublicKey:           publicKey,
			EncryptedPrivateKey: encryptedPrivateKey,
		},
		PlanType: 0,
	}

	orgCreationRequestBytes, err := json.Marshal(orgCreationRequest)
	if err != nil {
		return "", fmt.Errorf("unable to marshall organization creation request: %w", err)
	}

	req, err := http.NewRequest("POST", c.organizationURL(), bytes.NewBuffer(orgCreationRequestBytes))
	if err != nil {
		return "", fmt.Errorf("error preparing organization creation request: %w", err)
	}

	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.session.accessToken))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("device-type", c.deviceType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling organization creation: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading organization creation response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status code for organization creation call: %d!=200, body:%s", resp.StatusCode, string(body))
	}

	var orgCreationResponse CreateOrganizationResponse
	err = json.Unmarshal(body, &orgCreationResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling organization creation response: %w", err)
	}

	return orgCreationResponse.Id, nil
}

func (c *client) GetCollections(orgID string) (string, error) {
	req, err := http.NewRequest("GET", c.organizationCollectionURL(orgID), nil)
	if err != nil {
		return "", fmt.Errorf("error preparing collection retrieval request: %w", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", c.session.accessToken))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("device-type", c.deviceType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling collection retrieval: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading collection retrieval response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status code for collection retrieval call: %d!=200, body:%s", resp.StatusCode, string(body))
	}

	var collResponse CollectionResponse
	err = json.Unmarshal(body, &collResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling collection retrieval response: %w", err)
	}

	return collResponse.Data[0].Id, nil
}

func (c *client) signupURL() string       { return fmt.Sprintf("%s/api/accounts/register", c.serverURL) }
func (c *client) loginURL() string        { return fmt.Sprintf("%s/identity/connect/token", c.serverURL) }
func (c *client) organizationURL() string { return fmt.Sprintf("%s/api/organizations", c.serverURL) }
func (c *client) organizationCollectionURL(orgID string) string {
	return fmt.Sprintf("%s/api/organizations/%s/collections", c.serverURL, orgID)
}
