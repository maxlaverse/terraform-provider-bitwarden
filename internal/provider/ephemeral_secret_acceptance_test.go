//go:build offline

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const ephemeralCredential = "ephemeral-credential-must-not-be-persisted"

func TestEphemeralSecretProviderConfiguration(t *testing.T) {
	secret := models.Secret{
		ID: "secret-id", Key: "credential", Value: ephemeralCredential,
		Note:           "ephemeral-note-must-not-be-persisted",
		OrganizationID: "organization-id", ProjectID: "project-id",
	}
	for _, selector := range []string{`id = "secret-id"`, `key = "credential"`} {
		t.Run(selector, func(t *testing.T) {
			check := ephemeralSecretPersistenceCheck{}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralCredentialFactories(nil, secret),
				Steps: []resource.TestStep{{
					Config:            ephemeralSecretConsumerConfig(selector),
					ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}},
					ConfigStateChecks: []statecheck.StateCheck{check},
					Check:             resource.TestCheckOutput("project_name", "public-project"),
				}},
			})
		})
	}
}

func TestEphemeralSecretErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		client bitwarden.SecretsManager
		want   string
	}{
		{name: "no Secrets Manager client", want: "Provider not configured for Secrets Manager"},
		{name: "server error details", client: &ephemeralSecretErrorClient{err: &webapi.HTTPError{StatusCode: 503, Message: "Service temporarily unavailable"}}, want: "Service temporarily unavailable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralCredentialFactories(&ProviderClients{SecretsManager: tt.client}, models.Secret{}),
				Steps: []resource.TestStep{{
					Config:      ephemeralSecretConsumerConfig(`key = "credential"`),
					ExpectError: regexp.MustCompile(tt.want),
				}},
			})
		})
	}
}

func ephemeralSecretConsumerConfig(selector string) string {
	return fmt.Sprintf(`
provider "bitwarden" {
  alias        = "source"
  access_token = "bootstrap"
}

ephemeral "bitwarden_secret" "credential" {
  provider = bitwarden.source
  %s
}

provider "bitwarden" {
  alias        = "consumer"
  access_token = jsonencode(ephemeral.bitwarden_secret.credential)
}

data "bitwarden_project" "probe" {
  provider = bitwarden.consumer
  id       = "project-id"
}

output "project_name" {
  value = data.bitwarden_project.probe.name
}
`, selector)
}

func ephemeralCredentialFactories(clients *ProviderClients, secret models.Secret) map[string]func() (tfprotov6.ProviderServer, error) {
	credential, _ := json.Marshal(map[string]string{
		"id": secret.ID, "key": secret.Key, "value": secret.Value,
		"note": secret.Note, "organization_id": secret.OrganizationID, "project_id": secret.ProjectID,
	})
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bitwarden": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(&ephemeralCredentialProvider{
				bitwardenProvider: &bitwardenProvider{},
				sourceClients:     clients,
				secret:            secret,
				credential:        string(credential),
			})(), nil
		},
	}
}

// Use the real schemas and resource implementations, replacing only backend
// access so the CLI can exercise ephemeral provider configuration offline.
type ephemeralCredentialProvider struct {
	*bitwardenProvider
	sourceClients *ProviderClients
	secret        models.Secret
	credential    string
}

func (p *ephemeralCredentialProvider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	var model bitwardenProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clients := &ProviderClients{SecretsManager: &ephemeralCredentialClient{
		accessToken: model.AccessToken.ValueString(), secret: p.secret, credential: p.credential,
	}}
	if model.AccessToken.ValueString() == "bootstrap" && p.sourceClients != nil {
		clients = p.sourceClients
	}
	resp.ResourceData = clients
	resp.DataSourceData = clients
	resp.EphemeralResourceData = clients
}

type ephemeralCredentialClient struct {
	bitwarden.SecretsManager
	accessToken string
	secret      models.Secret
	credential  string
}

func (c *ephemeralCredentialClient) GetSecret(_ context.Context, secret models.Secret) (*models.Secret, error) {
	if secret.ID != c.secret.ID {
		return nil, models.ErrObjectNotFound
	}
	return &c.secret, nil
}

func (c *ephemeralCredentialClient) GetSecretByKey(_ context.Context, key string) (*models.Secret, error) {
	if key != c.secret.Key {
		return nil, models.ErrNoObjectFoundMatchingFilter
	}
	return &c.secret, nil
}

func (c *ephemeralCredentialClient) GetProject(_ context.Context, project models.Project) (*models.Project, error) {
	if c.accessToken != c.credential {
		return nil, fmt.Errorf("consumer did not receive the ephemeral credential")
	}
	return &models.Project{ID: project.ID, Name: "public-project", OrganizationID: "organization-id"}, nil
}

type ephemeralSecretErrorClient struct {
	bitwarden.SecretsManager
	err error
}

func (c *ephemeralSecretErrorClient) GetSecretByKey(context.Context, string) (*models.Secret, error) {
	return nil, c.err
}

type ephemeralSecretPersistenceCheck struct{}

func (ephemeralSecretPersistenceCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	resp.Error = checkEphemeralSecretPersistence(req.Plan)
}

func (ephemeralSecretPersistenceCheck) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	resp.Error = checkEphemeralSecretPersistence(req.State)
}

func checkEphemeralSecretPersistence(value any) error {
	content, err := json.Marshal(value)
	if err != nil {
		return err
	}
	for _, marker := range []string{ephemeralCredential, "ephemeral-note-must-not-be-persisted"} {
		if strings.Contains(string(content), marker) {
			return fmt.Errorf("ephemeral secret data was persisted")
		}
	}
	return nil
}
