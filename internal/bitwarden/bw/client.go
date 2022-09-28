package bw

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/executor"
)

var (
	ErrNotFound = errNotFound()
)

func errNotFound() error { return errors.New("resource not found") }

type Client interface {
	CreateObject(Object) (*Object, error)
	EditObject(Object) (*Object, error)
	HasSessionKey() bool
	SetSessionKey(string)
	GetObject(Object) (*Object, error)
	LoginWithPassword(username, password string) error
	LoginWithAPIKey(password, clientId, clientSecret string) error
	Logout() error
	RemoveObject(Object) error
	SetServer(string) error
	Status() (*Status, error)
	Sync() error
	Unlock(password string) error
}

func NewClient(execPath string, opts ...Options) Client {
	c := &client{
		execPath: execPath,
		executor: executor.DefaultExecutor,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

type client struct {
	appDataDir  string
	disableSync bool
	execPath    string
	executor    executor.Executor
	sessionKey  string
}

type Options func(c Client)

func WithAppDataDir(appDataDir string) Options {
	return func(c Client) {
		c.(*client).appDataDir = appDataDir
	}
}

func DisableSync() Options {
	return func(c Client) {
		c.(*client).disableSync = true
	}
}

func (c *client) CreateObject(obj Object) (*Object, error) {
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	out, err := c.cmdWithSession("create", string(obj.Object), objEncoded).RunCaptureOutput()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, unmarshallError("create object", err, out)
	}
	err = c.Sync()
	if err != nil {
		return nil, fmt.Errorf("error syncing: %v, %v", err, string(out))
	}
	return &obj, nil
}

func (c *client) EditObject(obj Object) (*Object, error) {
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	out, err := c.cmdWithSession("edit", string(obj.Object), obj.ID, objEncoded).RunCaptureOutput()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, unmarshallError("edit object", err, out)
	}
	err = c.Sync()
	if err != nil {
		return nil, fmt.Errorf("error syncing: %v, %v", err, string(out))
	}

	return &obj, nil
}

func (c *client) GetObject(obj Object) (*Object, error) {
	out, err := c.cmdWithSession("get", string(obj.Object), obj.ID).RunCaptureOutput()
	if err != nil {
		if string(out) == "Not found." {
			return nil, ErrNotFound
		}
		return nil, err
	}

	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, unmarshallError("get object", err, out)
	}

	return &obj, nil
}

// LoginWithPassword logs in using a password and retrieves the session key,
// allowing authenticated requests using the client.
func (c *client) LoginWithPassword(username, password string) error {
	out, err := c.cmd("login", username, "--raw", "--passwordenv", "BW_PASSWORD").WithEnv([]string{fmt.Sprintf("BW_PASSWORD=%s", password)}).RunCaptureOutput()
	if err != nil {
		return err
	}
	c.sessionKey = string(out)
	return nil
}

// LoginWithPassword logs in using an API key and unlock the Vault in order to retrieve a session key,
// allowing authenticated requests using the client.
func (c *client) LoginWithAPIKey(password, clientId, clientSecret string) error {
	err := c.cmd("login", "--apikey").WithEnv([]string{fmt.Sprintf("BW_CLIENTID=%s", clientId), fmt.Sprintf("BW_CLIENTSECRET=%s", clientSecret)}).Run()
	if err != nil {
		return err
	}
	return c.Unlock(password)
}

func (c *client) Logout() error {
	return c.cmd("logout").Run()
}

func (c *client) RemoveObject(obj Object) error {
	return c.cmdWithSession("delete", string(obj.Object), obj.ID).Run()
}

func (c *client) SetServer(server string) error {
	return c.cmd("config", "server", server).Run()
}

func (c *client) Status() (*Status, error) {
	out, err := c.cmdWithSession("status").RunCaptureOutput()
	if err != nil {
		return nil, err
	}

	var status Status
	err = json.Unmarshal(out, &status)
	if err != nil {
		return nil, unmarshallError("status", err, out)
	}

	return &status, nil
}

func (c *client) Unlock(password string) error {
	out, err := c.cmd("unlock", "--raw", "--passwordenv", "BW_PASSWORD").WithEnv([]string{fmt.Sprintf("BW_PASSWORD=%s", password)}).RunCaptureOutput()
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
	return c.cmdWithSession("sync").Run()
}

func (c *client) cmd(args ...string) executor.Command {
	return c.executor.NewCommand(c.execPath, args...).WithEnv(c.env())
}

func (c *client) cmdWithSession(args ...string) executor.Command {
	return c.cmd(args...).WithEnv([]string{fmt.Sprintf("BW_SESSION=%s", c.sessionKey)})
}

func (c *client) env() []string {
	return append(
		os.Environ(),
		fmt.Sprintf("BITWARDENCLI_APPDATA_DIR=%s", c.appDataDir),
		"BW_NOINTERACTION=true",
	)
}

func (c *client) encode(item Object) (string, error) {
	newOut, err := json.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("marshalling error: %v, %v", err, string(newOut))
	}

	out, err := c.cmd("encode").WithStdin(string(newOut)).RunCaptureOutput()
	if err != nil {
		return "", fmt.Errorf("encoding error: %v, %v", err, string(newOut))
	}
	return string(out), err
}

func unmarshallError(cmd string, err error, out []byte) error {
	return fmt.Errorf("unable to parse '%s' result: %v, output: %v", cmd, err, string(out))
}
