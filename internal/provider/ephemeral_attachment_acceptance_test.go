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

const ephemeralAttachmentCredential = "ephemeral-attachment-content-must-not-be-persisted"

func TestEphemeralAttachmentProviderConfiguration(t *testing.T) {
	check := ephemeralAttachmentPersistenceCheck{}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ephemeralAttachmentFactories(nil, "item-id", "attachment-id", ephemeralAttachmentCredential),
		Steps: []resource.TestStep{{
			Config:            ephemeralAttachmentConsumerConfig(`id = "attachment-id"` + "\n  item_id = \"item-id\""),
			ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}},
			ConfigStateChecks: []statecheck.StateCheck{check},
			Check:             resource.TestCheckOutput("folder_name", "public-folder"),
		}},
	})
}

func TestEphemeralAttachmentErrors(t *testing.T) {
	for _, tt := range []struct {
		name     string
		clients  *ProviderClients
		selector string
		want     string
	}{
		{name: "no Password Manager client", clients: &ProviderClients{}, selector: `id = "attachment-id"` + "\n  item_id = \"item-id\"", want: "Provider not configured for Password Manager"},
		{name: "server error details", clients: &ProviderClients{PasswordManager: &ephemeralAttachmentErrorClient{err: &webapi.HTTPError{StatusCode: 503, Message: "Service temporarily unavailable"}}}, selector: `id = "attachment-id"` + "\n  item_id = \"item-id\"", want: "Service temporarily unavailable"},
		{name: "object not found", clients: &ProviderClients{PasswordManager: &ephemeralAttachmentClient{itemID: "item-id", id: "attachment-id", content: ephemeralAttachmentCredential}}, selector: `id = "missing-attachment"` + "\n  item_id = \"item-id\"", want: "object not found"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralAttachmentFactories(tt.clients, "item-id", "attachment-id", ephemeralAttachmentCredential),
				Steps: []resource.TestStep{{
					Config:      ephemeralAttachmentConsumerConfig(tt.selector),
					ExpectError: regexp.MustCompile(tt.want),
				}},
			})
		})
	}
}

func ephemeralAttachmentConsumerConfig(selector string) string {
	return fmt.Sprintf(`
provider "bitwarden" {
  alias       = "source"
  email       = "source@example.com"
  session_key = "bootstrap"
}

ephemeral "bitwarden_attachment" "credential" {
  provider = bitwarden.source
  %s
}

provider "bitwarden" {
  alias       = "consumer"
  email       = "consumer@example.com"
  session_key = ephemeral.bitwarden_attachment.credential.content
}

data "bitwarden_folder" "probe" {
  provider = bitwarden.consumer
  id       = "folder-id"
}

output "folder_name" {
  value = data.bitwarden_folder.probe.name
}
`, selector)
}

func ephemeralAttachmentFactories(clients *ProviderClients, itemID, id, content string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bitwarden": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(&ephemeralAttachmentProvider{
				bitwardenProvider: &bitwardenProvider{},
				sourceClients:     clients,
				itemID:            itemID,
				id:                id,
				credential:        content,
			})(), nil
		},
	}
}

// Use the real schemas and resource implementations, replacing only backend
// access so the CLI can exercise ephemeral provider configuration offline.
type ephemeralAttachmentProvider struct {
	*bitwardenProvider
	sourceClients *ProviderClients
	itemID        string
	id            string
	credential    string
}

func (p *ephemeralAttachmentProvider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	var model bitwardenProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clients := &ProviderClients{PasswordManager: &ephemeralAttachmentClient{
		sessionKey: model.SessionKey.ValueString(), itemID: p.itemID, id: p.id, content: p.credential, credential: p.credential,
	}}
	if model.SessionKey.ValueString() == "bootstrap" && p.sourceClients != nil {
		clients = p.sourceClients
	}
	resp.ResourceData = clients
	resp.DataSourceData = clients
	resp.EphemeralResourceData = clients
}

type ephemeralAttachmentClient struct {
	bitwarden.PasswordManager
	sessionKey string
	itemID     string
	id         string
	content    string
	credential string
}

func (c *ephemeralAttachmentClient) GetAttachment(_ context.Context, itemId, attachmentId string) ([]byte, error) {
	if itemId != c.itemID || attachmentId != c.id {
		return nil, models.ErrObjectNotFound
	}
	return []byte(c.content), nil
}

func (c *ephemeralAttachmentClient) GetFolder(_ context.Context, folder models.Folder) (*models.Folder, error) {
	if c.sessionKey != c.credential {
		return nil, fmt.Errorf("consumer did not receive the ephemeral credential")
	}
	return &models.Folder{ID: folder.ID, Name: "public-folder"}, nil
}

type ephemeralAttachmentErrorClient struct {
	bitwarden.PasswordManager
	err error
}

func (c *ephemeralAttachmentErrorClient) GetAttachment(context.Context, string, string) ([]byte, error) {
	return nil, c.err
}

type ephemeralAttachmentPersistenceCheck struct{}

func (ephemeralAttachmentPersistenceCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	resp.Error = checkEphemeralAttachmentPersistence(req.Plan)
}

func (ephemeralAttachmentPersistenceCheck) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	resp.Error = checkEphemeralAttachmentPersistence(req.State)
}

func checkEphemeralAttachmentPersistence(value any) error {
	content, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), ephemeralAttachmentCredential) {
		return fmt.Errorf("ephemeral attachment data was persisted")
	}
	return nil
}
