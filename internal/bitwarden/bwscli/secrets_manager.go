package bwscli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/command"
)

type SecretsManagerClient interface {
	CreateProject(ctx context.Context, project models.Project) (*models.Project, error)
	CreateSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	DeleteProject(ctx context.Context, project models.Project) error
	DeleteSecret(ctx context.Context, secret models.Secret) error
	EditProject(ctx context.Context, project models.Project) (*models.Project, error)
	EditSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	GetProject(ctx context.Context, project models.Project) (*models.Project, error)
	GetSecret(ctx context.Context, secret models.Secret) (*models.Secret, error)
	GetSecretByKey(ctx context.Context, secretKey string) (*models.Secret, error)
	LoginWithAccessToken(ctx context.Context, accessToken string) error
}

func NewSecretsManagerClient(serverURL string, opts ...Options) SecretsManagerClient {
	c := &client{
		serverURL: serverURL,
	}

	for _, o := range opts {
		o(c)
	}

	c.newCommand = command.New

	return c
}

type client struct {
	disableRetryBackoff bool
	serverURL           string
	newCommand          command.NewFn
	accessToken         string
}

type Options func(c bitwarden.SecretsManager)

func DisableRetryBackoff() Options {
	return func(c bitwarden.SecretsManager) {
		c.(*client).disableRetryBackoff = true
	}
}

func (c *client) LoginWithAccessToken(ctx context.Context, accessToken string) error {
	c.accessToken = accessToken
	return nil
}

func (c *client) CreateProject(ctx context.Context, project models.Project) (*models.Project, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"project",
		"create",
		project.Name,
	}

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var projectObj models.Project
	err = json.Unmarshal(out, &projectObj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	return &projectObj, nil
}

func (c *client) CreateSecret(ctx context.Context, secret models.Secret) (*models.Secret, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"secret",
		"create",
		secret.Key,
		secret.Value,
		secret.ProjectID,
	}

	if secret.Note != "" {
		args = append(args, "--note", secret.Note)
	}

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var secretObj models.Secret
	err = json.Unmarshal(out, &secretObj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	return &secretObj, nil
}

func (c *client) EditProject(ctx context.Context, project models.Project) (*models.Project, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"project",
		"edit",
		"--name", project.Name,
		project.ID,
	}

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var projectObj models.Project
	err = json.Unmarshal(out, &projectObj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	return &projectObj, nil
}

func (c *client) EditSecret(ctx context.Context, secret models.Secret) (*models.Secret, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"secret",
		"edit",
	}

	if secret.Key != "" {
		args = append(args, "--key", secret.Key)
	}
	if secret.Value != "" {
		args = append(args, "--value", secret.Value)
	}
	if secret.Note != "" {
		args = append(args, "--note", secret.Note)
	}
	if secret.ProjectID != "" {
		args = append(args, "--project-id", secret.ProjectID)
	}

	args = append(args, secret.ID)

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var secretObj models.Secret
	err = json.Unmarshal(out, &secretObj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	return &secretObj, nil
}

func (c *client) DeleteProject(ctx context.Context, project models.Project) error {
	if err := c.checkAccessToken(); err != nil {
		return err
	}

	_, err := c.cmdWithAccessToken("project", "delete", project.ID).Run(ctx)
	return remapError(err)
}

func (c *client) DeleteSecret(ctx context.Context, secret models.Secret) error {
	if err := c.checkAccessToken(); err != nil {
		return err
	}

	_, err := c.cmdWithAccessToken("secret", "delete", secret.ID).Run(ctx)
	return remapError(err)
}

func (c *client) GetProject(ctx context.Context, project models.Project) (*models.Project, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"project",
		"get",
		project.ID,
	}

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var projectObj models.Project
	err = json.Unmarshal(out, &projectObj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	return &projectObj, nil
}

func (c *client) GetSecret(ctx context.Context, secret models.Secret) (*models.Secret, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"secret",
		"get",
		secret.ID,
	}

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var secretObj models.Secret
	err = json.Unmarshal(out, &secretObj)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	return &secretObj, nil
}

func (c *client) GetSecretByKey(ctx context.Context, secretKey string) (*models.Secret, error) {
	if err := c.checkAccessToken(); err != nil {
		return nil, err
	}

	args := []string{
		"secret",
		"list",
	}

	out, err := c.cmdWithAccessToken(args...).Run(ctx)
	if err != nil {
		return nil, remapError(err)
	}

	var secrets []models.Secret
	err = json.Unmarshal(out, &secrets)
	if err != nil {
		return nil, newUnmarshallError(err, args[0:2], out)
	}

	matchingSecrets := []models.Secret{}
	for _, secret := range secrets {
		if secret.Key == secretKey {
			matchingSecrets = append(matchingSecrets, secret)
		}
	}

	if len(matchingSecrets) > 1 {
		return nil, models.ErrTooManyObjectsFound
	}
	if len(matchingSecrets) == 1 {
		return &matchingSecrets[0], nil
	}

	return nil, models.ErrNoObjectFoundMatchingFilter
}

func (c *client) cmdWithAccessToken(args ...string) command.Command {
	return c.newCommand("bws", args...).AppendEnv(c.env())
}

func (c *client) checkAccessToken() error {
	if c.accessToken == "" {
		return fmt.Errorf("access token not set")
	}
	return nil
}

func (c *client) env() []string {
	defaultEnv := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		"RUST_BACKTRACE=full",
		fmt.Sprintf("BWS_ACCESS_TOKEN=%s", c.accessToken),
		fmt.Sprintf("BWS_SERVER_URL=%s", c.serverURL),
	}
	return defaultEnv
}
