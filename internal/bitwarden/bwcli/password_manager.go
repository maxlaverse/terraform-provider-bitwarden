package bwcli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

type PasswordManagerClient interface {
	CreateAttachmentFromFile(ctx context.Context, itemId, filePath string) (*models.Object, error)
	CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Object, error)
	CreateObject(context.Context, models.Object) (*models.Object, error)
	EditObject(context.Context, models.Object) (*models.Object, error)
	GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error)
	GetObject(context.Context, models.Object) (*models.Object, error)
	GetSessionKey() string
	HasSessionKey() bool
	ListObjects(ctx context.Context, objType models.ObjectType, options ...bitwarden.ListObjectsOption) ([]models.Object, error)
	LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error
	LoginWithPassword(ctx context.Context, username, password string) error
	Logout(context.Context) error
	DeleteAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteObject(context.Context, models.Object) error
	SetServer(context.Context, string) error
	SetSessionKey(string)
	Status(context.Context) (*Status, error)
	Sync(context.Context) error
	Unlock(ctx context.Context, password string) error
}

func NewPasswordManagerClient(execPath string, opts ...Options) PasswordManagerClient {
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

type Options func(c bitwarden.PasswordManager)

func WithAppDataDir(appDataDir string) Options {
	return func(c bitwarden.PasswordManager) {
		c.(*client).appDataDir = appDataDir
	}
}

func WithExtraCACertsPath(extraCACertsPath string) Options {
	return func(c bitwarden.PasswordManager) {
		c.(*client).extraCACertsPath = extraCACertsPath
	}
}

func DisableSync() Options {
	return func(c bitwarden.PasswordManager) {
		c.(*client).disableSync = true
	}
}

func DisableRetryBackoff() Options {
	return func(c bitwarden.PasswordManager) {
		c.(*client).disableRetryBackoff = true
	}
}

func (c *client) CreateObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	obj.Groups = []interface{}{}
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	args := []string{
		"create",
		string(obj.Object),
		objEncoded,
	}

	if obj.Object == models.ObjectTypeOrgCollection {
		args = append(args, "--organizationid", obj.OrganizationID)
	}

	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	// NOTE(maxime): there is no need to sync after creating an item
	// as the creation issued an API call on the Vault directly.
	return &obj, nil
}

func (c *client) CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Object, error) {
	return nil, fmt.Errorf("creating attachments from content is only supported by the embedded client")
}

func (c *client) CreateAttachmentFromFile(ctx context.Context, itemId string, filePath string) (*models.Object, error) {
	out, err := c.cmdWithSession("create", string(models.ObjectTypeAttachment), "--itemid", itemId, "--file", filePath).Run(ctx)
	if err != nil {
		return nil, err
	}

	var obj models.Object
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, err
	}

	// NOTE(maxime): there is no need to sync after creating an item
	// as the creation issued an API call on the Vault directly.
	return &obj, nil
}

func (c *client) EditObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	obj.Groups = []interface{}{}
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	args := []string{
		"edit",
		string(obj.Object),
		obj.ID,
	}

	if obj.Object == models.ObjectTypeOrgCollection {
		args = append(args, "--organizationid", obj.OrganizationID)
	}
	args = append(args, []string{objEncoded}...)

	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}
	err = c.Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("error syncing: %v, %v", err, string(out))
	}

	return &obj, nil
}

func (c *client) GetObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	args := []string{
		"get",
		string(obj.Object),
		obj.ID,
	}

	desiredObjType := obj.Type

	if obj.Object == models.ObjectTypeOrgCollection {
		args = append(args, "--organizationid", obj.OrganizationID)
	}

	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	if desiredObjType > 0 && obj.Type != desiredObjType {
		return nil, models.ErrItemTypeMismatch
	}

	return &obj, nil
}

func (c *client) GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error) {
	out, err := c.cmdWithSession("get", string(models.ObjectTypeAttachment), attachmentId, "--itemid", itemId, "--raw").Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	return out, nil
}

func (c *client) GetSessionKey() string {
	return c.sessionKey
}

// ListObjects returns objects of a given type matching given filters.
func (c *client) ListObjects(ctx context.Context, objType models.ObjectType, options ...bitwarden.ListObjectsOption) ([]models.Object, error) {
	args := []string{
		"list",
		fmt.Sprintf("%ss", objType),
	}

	applyFiltersToArgs(&args, options...)

	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var objs []models.Object
	err = json.Unmarshal(out, &objs)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	filters := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	filteredObj := []models.Object{}
	for _, obj := range objs {
		if filters.ItemType != 0 && obj.Type != filters.ItemType {
			continue
		}
		filteredObj = append(filteredObj, obj)
	}

	return filteredObj, nil
}

// LoginWithPassword logs in using a password and retrieves the session key,
// allowing authenticated requests using the client.
func (c *client) LoginWithPassword(ctx context.Context, username, password string) error {
	out, err := c.cmd("login", username, "--raw", "--passwordenv", "BW_PASSWORD").AppendEnv([]string{fmt.Sprintf("BW_PASSWORD=%s", password)}).Run(ctx)
	if err != nil {
		return err
	}
	c.sessionKey = string(out)
	return nil
}

// LoginWithPassword logs in using an API key and unlock the Vault in order to retrieve a session key,
// allowing authenticated requests using the client.
func (c *client) LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error {
	_, err := c.cmd("login", "--apikey").AppendEnv([]string{fmt.Sprintf("BW_CLIENTID=%s", clientId), fmt.Sprintf("BW_CLIENTSECRET=%s", clientSecret)}).Run(ctx)
	if err != nil {
		return err
	}
	return c.Unlock(ctx, password)
}

func (c *client) Logout(ctx context.Context) error {
	_, err := c.cmd("logout").Run(ctx)
	return err
}

func (c *client) DeleteObject(ctx context.Context, obj models.Object) error {
	args := []string{
		"delete",
		string(obj.Object),
		obj.ID,
	}

	if obj.Object == models.ObjectTypeOrgCollection {
		args = append(args, "--organizationid", obj.OrganizationID)
	}

	_, err := c.cmdWithSession(args...).Run(ctx)
	return err
}

func (c *client) DeleteAttachment(ctx context.Context, itemId, attachmentId string) error {
	// TODO: Don't fail if attachment is already gone
	_, err := c.cmdWithSession("delete", string(models.ObjectTypeAttachment), attachmentId, "--itemid", itemId).Run(ctx)
	return err
}

func (c *client) SetServer(ctx context.Context, server string) error {
	_, err := c.cmd("config", "server", server).Run(ctx)
	return err
}

func (c *client) Status(ctx context.Context) (*Status, error) {
	out, err := c.cmdWithSession("status").Run(ctx)
	if err != nil {
		return nil, err
	}

	var status Status
	err = json.Unmarshal(out, &status)
	if err != nil {
		return nil, newUnmarshallError(err, []string{"status"}, out)
	}

	return &status, nil
}

func (c *client) Unlock(ctx context.Context, password string) error {
	out, err := c.cmd("unlock", "--raw", "--passwordenv", "BW_PASSWORD").AppendEnv([]string{fmt.Sprintf("BW_PASSWORD=%s", password)}).Run(ctx)
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

func (c *client) Sync(ctx context.Context) error {
	if c.disableSync {
		return nil
	}
	_, err := c.cmdWithSession("sync").Run(ctx)
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

func (c *client) encode(item models.Object) (string, error) {
	newOut, err := json.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("marshalling error: %v, %v", err, string(newOut))
	}
	return base64.RawStdEncoding.EncodeToString(newOut), nil
}

func applyFiltersToArgs(args *[]string, options ...bitwarden.ListObjectsOption) {
	filters := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if filters.OrganizationFilter != "" {
		*args = append(*args, "--organizationid", filters.OrganizationFilter)
	}
	if filters.FolderFilter != "" {
		*args = append(*args, "--folderid", filters.FolderFilter)
	}
	if filters.CollectionFilter != "" {
		*args = append(*args, "--collectionid", filters.CollectionFilter)
	}
	if filters.SearchFilter != "" {
		*args = append(*args, "--search", filters.SearchFilter)
	}
	if filters.UrlFilter != "" {
		*args = append(*args, "--url", filters.UrlFilter)
	}
}
