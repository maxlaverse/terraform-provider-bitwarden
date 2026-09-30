//go:build offline

package provider

import (
	"context"
	"encoding/json"
	"fmt"
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
)

const ephemeralCredential = "ephemeral-credential-must-not-be-persisted"

func TestEphemeralSecretProviderConfiguration(t *testing.T) {
	testEphemeralSecretProviderConfiguration(t, "bitwarden_secret", `id = "secret-id"`, "ephemeral.bitwarden_secret.credential.value")
}

func testEphemeralSecretProviderConfiguration(t *testing.T, resourceType, selector, credentialExpression string) {
	t.Helper()
	check := ephemeralSecretPersistenceCheck{}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"bitwarden": func() (tfprotov6.ProviderServer, error) {
				return providerserver.NewProtocol6(&ephemeralCredentialProvider{
					bitwardenProvider: &bitwardenProvider{},
				})(), nil
			},
		},
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
provider "bitwarden" {
  alias        = "source"
  access_token = "bootstrap"
}

ephemeral %q "credential" {
  provider = bitwarden.source
  %s
}

provider "bitwarden" {
  alias        = "consumer"
  access_token = %s
}

data "bitwarden_project" "probe" {
  provider = bitwarden.consumer
  id       = "project-id"
}

output "project_name" {
  value = data.bitwarden_project.probe.name
}
`, resourceType, selector, credentialExpression),
			ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}},
			ConfigStateChecks: []statecheck.StateCheck{check},
			Check:             resource.TestCheckOutput("project_name", "public-project"),
		}},
	})
}

// Use the real schemas and resource implementations, replacing only backend
// access so the CLI can exercise ephemeral provider configuration offline.
type ephemeralCredentialProvider struct {
	*bitwardenProvider
}

func (p *ephemeralCredentialProvider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	var model bitwardenProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clients := &ProviderClients{SecretsManager: &ephemeralCredentialClient{accessToken: model.AccessToken.ValueString()}}
	resp.ResourceData = clients
	resp.DataSourceData = clients
	resp.EphemeralResourceData = clients
}

type ephemeralCredentialClient struct {
	bitwarden.SecretsManager
	accessToken string
}

func (c *ephemeralCredentialClient) GetSecret(_ context.Context, secret models.Secret) (*models.Secret, error) {
	return &models.Secret{ID: secret.ID, Key: "credential", Value: ephemeralCredential, Note: "ephemeral-note-must-not-be-persisted"}, nil
}

func (c *ephemeralCredentialClient) GetProject(_ context.Context, project models.Project) (*models.Project, error) {
	if c.accessToken != ephemeralCredential {
		return nil, fmt.Errorf("consumer did not receive the ephemeral credential")
	}
	return &models.Project{ID: project.ID, Name: "public-project", OrganizationID: "organization-id"}, nil
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
