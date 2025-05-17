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
	CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Item, error)
	CreateAttachmentFromFile(ctx context.Context, itemId, filePath string) (*models.Item, error)
	CreateFolder(context.Context, models.Folder) (*models.Folder, error)
	CreateItem(context.Context, models.Item) (*models.Item, error)
	CreateOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	DeleteAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteFolder(context.Context, models.Folder) error
	DeleteItem(context.Context, models.Item) error
	DeleteOrganizationCollection(ctx context.Context, obj models.OrgCollection) error
	EditFolder(context.Context, models.Folder) (*models.Folder, error)
	EditItem(context.Context, models.Item) (*models.Item, error)
	EditOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	FindFolder(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Folder, error)
	FindItem(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Item, error)
	FindOrganization(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Organization, error)
	FindOrganizationMember(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgMember, error)
	FindOrganizationCollection(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgCollection, error)
	GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error)
	GetFolder(context.Context, models.Folder) (*models.Folder, error)
	GetItem(context.Context, models.Item) (*models.Item, error)
	GetOrganization(context.Context, models.Organization) (*models.Organization, error)
	GetOrganizationMember(context.Context, models.OrgMember) (*models.OrgMember, error)
	GetOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	GetSessionKey() string
	HasSessionKey() bool
	LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error
	LoginWithPassword(ctx context.Context, username, password string) error
	Logout(context.Context) error
	SetServer(context.Context, string) error
	SetSessionKey(string)
	Status(context.Context) (*Status, error)
	Sync(context.Context) error
	Unlock(ctx context.Context, password string) error
}

func NewPasswordManagerClient(opts ...Options) PasswordManagerClient {
	c := &client{}

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

func (c *client) CreateAttachmentFromFile(ctx context.Context, itemId string, filePath string) (*models.Item, error) {
	out, err := c.cmdWithSession("create", string(models.ObjectTypeAttachment), "--itemid", itemId, "--file", filePath).Run(ctx)
	if err != nil {
		return nil, err
	}

	var obj models.Item
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, err
	}

	// NOTE(maxime): there is no need to sync after creating an item as the
	// creation issued an API call on the Vault directly which refreshes the
	// local cache.
	return &obj, nil
}

func (c *client) CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Item, error) {
	return nil, fmt.Errorf("creating attachments from content is only supported by the embedded client")
}

func (c *client) CreateFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	return createObject(ctx, c, obj, models.ObjectTypeFolder)
}

func (c *client) CreateItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	return createObject(ctx, c, obj, models.ObjectTypeItem)
}

func (c *client) CreateOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	obj.Groups = []interface{}{}
	if len(obj.Users) > 0 {
		return nil, fmt.Errorf("managing collection memberships is only supported by the embedded client")
	}
	return createObject(ctx, c, obj, models.ObjectTypeOrgCollection)
}

func createObject[T any](ctx context.Context, c *client, obj T, objectType models.ObjectType) (*T, error) {
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	args := []string{
		"create",
		string(objectType),
		objEncoded,
	}

	switch orgObj := any(obj).(type) {
	case models.OrgCollection:
		args = append(args, "--organizationid", orgObj.OrganizationID)
	}
	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	// NOTE(maxime): there is no need to sync after creating an item as the
	// creation issued an API call on the Vault directly which refreshes the
	// local cache.
	return &obj, nil
}

func (c *client) EditFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	return editGenericObject(ctx, c, obj, obj.Object, obj.ID)
}

func (c *client) EditItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	return editGenericObject(ctx, c, obj, obj.Object, obj.ID)
}

func (c *client) EditOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	obj.Groups = []interface{}{}
	if len(obj.Users) > 0 {
		return nil, fmt.Errorf("managing collection memberships is only supported by the embedded client")
	}
	return editGenericObject(ctx, c, obj, obj.Object, obj.ID)
}

func editGenericObject[T any](ctx context.Context, c *client, obj T, objectType models.ObjectType, id string) (*T, error) {
	objEncoded, err := c.encode(obj)
	if err != nil {
		return nil, err
	}

	args := []string{
		"edit",
		string(objectType),
		id,
	}

	switch orgObj := any(obj).(type) {
	case models.OrgCollection:
		args = append(args, "--organizationid", orgObj.OrganizationID)
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

func (c *client) GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error) {
	out, err := c.cmdWithSession("get", string(models.ObjectTypeAttachment), attachmentId, "--itemid", itemId, "--raw").Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	return out, nil
}

func (c *client) GetFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	return getObject(ctx, c, obj, obj.Object, obj.ID)
}

func (c *client) GetItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	return getObject(ctx, c, obj, obj.Object, obj.ID)
}

func (c *client) GetOrganization(ctx context.Context, obj models.Organization) (*models.Organization, error) {
	return getObject(ctx, c, obj, obj.Object, obj.ID)
}

func (c *client) GetOrganizationMember(ctx context.Context, obj models.OrgMember) (*models.OrgMember, error) {
	return nil, fmt.Errorf("getting organization members is only supported by the embedded client")
}

func (c *client) GetOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	return getObject(ctx, c, obj, obj.Object, obj.ID)
}

func getObject[T any](ctx context.Context, c *client, obj T, objectType models.ObjectType, id string) (*T, error) {
	args := []string{
		"get",
		string(objectType),
		id,
	}

	switch orgObj := any(obj).(type) {
	case models.OrgCollection:
		args = append(args, "--organizationid", orgObj.OrganizationID)
	}

	var desiredObjType models.ItemType

	switch itemObj := any(obj).(type) {
	case models.Item:
		desiredObjType = itemObj.Type
	}

	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	err = json.Unmarshal(out, &obj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	switch itemObj := any(obj).(type) {
	case models.Item:
		if desiredObjType > 0 && itemObj.Type != desiredObjType {
			return nil, models.ErrItemTypeMismatch
		}
	}

	return &obj, nil
}

func (c *client) GetSessionKey() string {
	return c.sessionKey
}

func (c *client) FindFolder(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Folder, error) {
	return findGenericObject[models.Folder](ctx, c, models.ObjectTypeFolder, options...)
}

func (c *client) FindItem(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Item, error) {
	return findGenericObject[models.Item](ctx, c, models.ObjectTypeItem, options...)
}

func (c *client) FindOrganization(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Organization, error) {
	return findGenericObject[models.Organization](ctx, c, models.ObjectTypeOrganization, options...)
}

func (c *client) FindOrganizationMember(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgMember, error) {
	return nil, fmt.Errorf("find organization members is only supported by the embedded client")
}

func (c *client) FindOrganizationCollection(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgCollection, error) {
	return findGenericObject[models.OrgCollection](ctx, c, models.ObjectTypeOrgCollection, options...)
}

func findGenericObject[T any](ctx context.Context, c *client, objType models.ObjectType, options ...bitwarden.ListObjectsOption) (*T, error) {
	args := []string{
		"list",
		fmt.Sprintf("%ss", objType),
	}

	applyFiltersToArgs(&args, options...)

	out, err := c.cmdWithSession(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var foundObjects []T
	err = json.Unmarshal(out, &foundObjects)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	filters := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	filteredObj := []T{}
	for _, obj := range foundObjects {
		switch itemObj := any(obj).(type) {
		case models.Item:
			if filters.ItemType > 0 && itemObj.Type != filters.ItemType {
				continue
			}
		}
		filteredObj = append(filteredObj, obj)
	}

	if len(filteredObj) == 0 {
		return nil, models.ErrNoObjectFoundMatchingFilter
	} else if len(filteredObj) > 1 {
		return nil, models.ErrTooManyObjectsFound
	}

	return &filteredObj[0], nil
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

func (c *client) DeleteFolder(ctx context.Context, obj models.Folder) error {
	_, err := c.cmdWithSession("delete", string(models.ObjectTypeFolder), obj.ID).Run(ctx)
	return err
}

func (c *client) DeleteItem(ctx context.Context, obj models.Item) error {
	_, err := c.cmdWithSession("delete", string(models.ObjectTypeItem), obj.ID).Run(ctx)
	return err
}

func (c *client) DeleteOrganizationCollection(ctx context.Context, obj models.OrgCollection) error {
	_, err := c.cmdWithSession("delete", string(models.ObjectTypeOrgCollection), obj.ID, "--organizationid", obj.OrganizationID).Run(ctx)
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
	return c.newCommand("bw", args...).AppendEnv(c.env())
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

func (c *client) encode(item interface{}) (string, error) {
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
