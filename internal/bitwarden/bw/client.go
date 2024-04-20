package bw

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

type Client interface {
	CreateAttachment(itemId, filePath string) (*Object, error)
	CreateObject(Object) (*Object, error)
	EditObject(Object) (*Object, error)
	GetAttachment(itemId, attachmentId string) ([]byte, error)
	GetObject(objType, itemOrSearch string) (*Object, error)
	GetSessionKey() string
	HasSessionKey() bool
	ListObjects(objType string, options ...ListObjectsOption) ([]Object, error)
	LoginWithAPIKey(password, clientId, clientSecret string) error
	LoginWithPassword(username, password string) error
	Logout() error
	DeleteAttachment(itemId, attachmentId string) error
	DeleteObject(objType, itemId string) error
	SetServer(string) error
	SetSessionKey(string)
	Status() (*Status, error)
	Sync() error
	Unlock(password string) error
}

func NewClient(execPath string, opts ...Options) Client {
	c := &client{
		execPath: execPath,
	}

	for _, o := range opts {
		o(c)
	}

	c.newCommand = command.NewWithRetries(&retryHandler{disableRetryBackoff: c.disableRetryBackoff})

	return c
}

type client struct {
	appDataDir          string
	disableSync         bool
	disableRetryBackoff bool
	execPath            string
	extraCACertsPath    string
	newCommand          command.NewFn
	sessionKey          string
}

type Options func(c Client)

func WithAppDataDir(appDataDir string) Options {
	return func(c Client) {
		c.(*client).appDataDir = appDataDir
	}
}

func WithExtraCACertsPath(extraCACertsPath string) Options {
	return func(c Client) {
		c.(*client).extraCACertsPath = extraCACertsPath
	}
}

func DisableSync() Options {
	return func(c Client) {
		c.(*client).disableSync = true
	}
}

func DisableRetryBackoff() Options {
	return func(c Client) {
		c.(*client).disableRetryBackoff = true
	}
}

func (c *client) CreateObject(obj Object) (*Object, error) {
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	out, err := c.cmdWithSession("create", string(obj.Object), objEncoded).Run()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, "create object", out)
	}

	// NOTE(maxime): there is no need to sync after creating an item
	// as the creation issued an API call on the Vault directly.
	return &obj, nil
}

func (c *client) CreateAttachment(itemId string, filePath string) (*Object, error) {
	out, err := c.cmdWithSession("create", string(ObjectTypeAttachment), "--itemid", itemId, "--file", filePath).Run()
	if err != nil {
		return nil, err
	}

	var obj Object
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, err
	}

	// NOTE(maxime): there is no need to sync after creating an item
	// as the creation issued an API call on the Vault directly.
	return &obj, nil
}

func (c *client) EditObject(obj Object) (*Object, error) {
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	out, err := c.cmdWithSession("edit", string(obj.Object), obj.ID, objEncoded).Run()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, "edit object", out)
	}
	err = c.Sync()
	if err != nil {
		return nil, fmt.Errorf("error syncing: %v, %v", err, string(out))
	}

	return &obj, nil
}

func (c *client) GetObject(objType, itemOrSearch string) (*Object, error) {
	out, err := c.cmdWithSession("get", objType, itemOrSearch).Run()
	if err != nil {
		return nil, remapError(err)
	}

	var obj Object
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, "get object", out)
	}

	return &obj, nil
}

func (c *client) GetAttachment(itemId, attachmentId string) ([]byte, error) {
	out, err := c.cmdWithSession("get", string(ObjectTypeAttachment), attachmentId, "--itemid", itemId, "--raw").Run()
	if err != nil {
		return nil, remapError(err)
	}

	return out, nil
}

func (c *client) GetSessionKey() string {
	return c.sessionKey
}

// ListObjects returns objects of a given type matching given filters.
func (c *client) ListObjects(objType string, options ...ListObjectsOption) ([]Object, error) {
	args := []string{
		"list",
		objType,
	}

	for _, applyOption := range options {
		applyOption(&args)
	}

	out, err := c.cmdWithSession(args...).Run()
	if err != nil {
		return nil, remapError(err)
	}

	var obj []Object
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, "list object", out)
	}

	return obj, nil
}

// LoginWithPassword logs in using a password and retrieves the session key,
// allowing authenticated requests using the client.
func (c *client) LoginWithPassword(username, password string) error {
	out, err := c.cmd("login", username, "--raw", "--passwordenv", "BW_PASSWORD").AppendEnv([]string{fmt.Sprintf("BW_PASSWORD=%s", password)}).Run()
	if err != nil {
		return err
	}
	c.sessionKey = string(out)
	return nil
}

// LoginWithPassword logs in using an API key and unlock the Vault in order to retrieve a session key,
// allowing authenticated requests using the client.
func (c *client) LoginWithAPIKey(password, clientId, clientSecret string) error {
	_, err := c.cmd("login", "--apikey").AppendEnv([]string{fmt.Sprintf("BW_CLIENTID=%s", clientId), fmt.Sprintf("BW_CLIENTSECRET=%s", clientSecret)}).Run()
	if err != nil {
		return err
	}
	return c.Unlock(password)
}

func (c *client) Logout() error {
	_, err := c.cmd("logout").Run()
	return err
}

func (c *client) DeleteObject(objType, itemId string) error {
	_, err := c.cmdWithSession("delete", objType, itemId).Run()
	return err
}

func (c *client) DeleteAttachment(itemId, attachmentId string) error {
	_, err := c.cmdWithSession("delete", string(ObjectTypeAttachment), attachmentId, "--itemid", itemId).Run()
	return err
}

func (c *client) SetServer(server string) error {
	_, err := c.cmd("config", "server", server).Run()
	return err
}

func (c *client) Status() (*Status, error) {
	out, err := c.cmdWithSession("status").Run()
	if err != nil {
		return nil, err
	}

	var status Status
	err = json.Unmarshal(out, &status)
	if err != nil {
		return nil, newUnmarshallError(err, "status", out)
	}

	return &status, nil
}

func (c *client) Unlock(password string) error {
	out, err := c.cmd("unlock", "--raw", "--passwordenv", "BW_PASSWORD").AppendEnv([]string{fmt.Sprintf("BW_PASSWORD=%s", password)}).Run()
	if err != nil {
		return err
	}

	c.sessionKey = string(out)
	return nil
}

func (c *client) HasSessionKey() bool {
	return len(c.sessionKey) > 0
}

func (c *client) SetSessionKey(sessionKey string) {
	c.sessionKey = sessionKey
}

func (c *client) Sync() error {
	if c.disableSync {
		return nil
	}
	_, err := c.cmdWithSession("sync").Run()
	return err
}

func (c *client) cmd(args ...string) command.Command {
	return c.newCommand(c.execPath, args...).AppendEnv(c.env())
}

func (c *client) cmdWithSession(args ...string) command.Command {
	return c.cmd(args...).AppendEnv([]string{fmt.Sprintf("BW_SESSION=%s", c.sessionKey)})
}

func (c *client) env() []string {
	defaultEnv := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("BITWARDENCLI_APPDATA_DIR=%s", c.appDataDir),
		"BW_NOINTERACTION=true",
	}
	if len(c.extraCACertsPath) > 0 {
		return append(defaultEnv, fmt.Sprintf("NODE_EXTRA_CA_CERTS=%s", c.extraCACertsPath))
	}
	return defaultEnv
}

func (c *client) encode(item Object) (string, error) {
	newOut, err := json.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("marshalling error: %v, %v", err, string(newOut))
	}

	out, err := c.cmd("encode").WithStdin(string(newOut)).Run()
	if err != nil {
		return "", fmt.Errorf("encoding error: %v, %v", err, string(newOut))
	}
	return string(out), err
}
