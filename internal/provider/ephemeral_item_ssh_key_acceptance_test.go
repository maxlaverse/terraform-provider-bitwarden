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

const ephemeralSSHCredential = "ephemeral-private-key-must-not-be-persisted"

func TestEphemeralItemSSHKeyProviderConfiguration(t *testing.T) {
	item := ephemeralSSHKeyItem()
	for _, tt := range []struct {
		name     string
		selector string
	}{
		{name: "by ID", selector: `id = "item-id"`},
		{name: "by search", selector: `search = "Git/SSH Key"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			check := ephemeralItemSSHKeyPersistenceCheck{}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralSSHKeyFactories(nil, item),
				Steps: []resource.TestStep{{
					Config:            ephemeralItemSSHKeyConsumerConfig(tt.selector),
					ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}},
					ConfigStateChecks: []statecheck.StateCheck{check},
					Check:             resource.TestCheckOutput("folder_name", "public-folder"),
				}},
			})
		})
	}
}

func TestEphemeralItemSSHKeyErrors(t *testing.T) {
	item := ephemeralSSHKeyItem()
	for _, tt := range []struct {
		name     string
		clients  *ProviderClients
		selector string
		want     string
	}{
		{name: "no Password Manager client", clients: &ProviderClients{}, selector: `id = "item-id"`, want: "Provider not configured for Password Manager"},
		{name: "server error details", clients: &ProviderClients{PasswordManager: &ephemeralSSHKeyErrorClient{err: &webapi.HTTPError{StatusCode: 503, Message: "Service temporarily unavailable"}}}, selector: `search = "Git/SSH Key"`, want: "Service temporarily unavailable"},
		{name: "object not found", clients: &ProviderClients{PasswordManager: &ephemeralSSHKeyClient{item: item}}, selector: `id = "missing-item"`, want: "object not found"},
		{name: "no object found matching filter", clients: &ProviderClients{PasswordManager: &ephemeralSSHKeyClient{item: item}}, selector: `search = "missing-item"`, want: "no object found matching the filter"},
		{name: "too many objects found", clients: &ProviderClients{PasswordManager: &ephemeralSSHKeyErrorClient{err: models.ErrTooManyObjectsFound}}, selector: `search = "Git/SSH Key"`, want: "too many objects found"},
		{name: "item type mismatch", clients: &ProviderClients{PasswordManager: &ephemeralSSHKeyErrorClient{err: models.ErrItemTypeMismatch}}, selector: `id = "item-id"`, want: "returned object type does not match requested object type"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralSSHKeyFactories(tt.clients, item),
				Steps: []resource.TestStep{{
					Config:      ephemeralItemSSHKeyConsumerConfig(tt.selector),
					ExpectError: regexp.MustCompile(tt.want),
				}},
			})
		})
	}
}

func ephemeralSSHKeyItem() models.Item {
	return models.Item{
		ID:     "item-id",
		Name:   "Git/SSH Key",
		Object: models.ObjectTypeItem,
		Type:   models.ItemTypeSSHKey,
		Notes:  "ephemeral-ssh-notes-must-not-be-persisted",
		SSHKey: models.SSHKey{
			PrivateKey:     ephemeralSSHCredential,
			PublicKey:      "ephemeral-public-key-must-not-be-persisted",
			KeyFingerprint: "ephemeral-fingerprint-must-not-be-persisted",
		},
	}
}

func ephemeralItemSSHKeyConsumerConfig(selector string) string {
	return fmt.Sprintf(`
provider "bitwarden" {
  alias       = "source"
  email       = "source@example.com"
  session_key = "bootstrap"
}

ephemeral "bitwarden_item_ssh_key" "credential" {
  provider = bitwarden.source
  %s
}

provider "bitwarden" {
  alias       = "consumer"
  email       = "consumer@example.com"
  session_key = ephemeral.bitwarden_item_ssh_key.credential.private_key
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

func ephemeralSSHKeyFactories(clients *ProviderClients, item models.Item) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bitwarden": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(&ephemeralSSHKeyProvider{
				bitwardenProvider: &bitwardenProvider{},
				sourceClients:     clients,
				item:              item,
				credential:        item.SSHKey.PrivateKey,
			})(), nil
		},
	}
}

type ephemeralSSHKeyProvider struct {
	*bitwardenProvider
	sourceClients *ProviderClients
	item          models.Item
	credential    string
}

func (p *ephemeralSSHKeyProvider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	var model bitwardenProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clients := &ProviderClients{PasswordManager: &ephemeralSSHKeyClient{
		sessionKey: model.SessionKey.ValueString(), item: p.item, credential: p.credential,
	}}
	if model.SessionKey.ValueString() == "bootstrap" && p.sourceClients != nil {
		clients = p.sourceClients
	}
	resp.ResourceData = clients
	resp.DataSourceData = clients
	resp.EphemeralResourceData = clients
}

type ephemeralSSHKeyClient struct {
	bitwarden.PasswordManager
	sessionKey string
	item       models.Item
	credential string
}

func (c *ephemeralSSHKeyClient) GetItem(_ context.Context, item models.Item) (*models.Item, error) {
	if item.ID != c.item.ID {
		return nil, models.ErrObjectNotFound
	}
	if item.Type != 0 && item.Type != c.item.Type {
		return nil, models.ErrItemTypeMismatch
	}
	return &c.item, nil
}

func (c *ephemeralSSHKeyClient) FindItem(_ context.Context, options ...bitwarden.ListObjectsOption) (*models.Item, error) {
	filter := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if filter.SearchFilter != "" && filter.SearchFilter != c.item.Name {
		return nil, models.ErrNoObjectFoundMatchingFilter
	}
	if filter.ItemType != 0 && filter.ItemType != c.item.Type {
		return nil, models.ErrNoObjectFoundMatchingFilter
	}
	return &c.item, nil
}

func (c *ephemeralSSHKeyClient) GetFolder(_ context.Context, folder models.Folder) (*models.Folder, error) {
	if c.sessionKey != c.credential {
		return nil, fmt.Errorf("consumer did not receive the ephemeral credential")
	}
	return &models.Folder{ID: folder.ID, Name: "public-folder"}, nil
}

type ephemeralSSHKeyErrorClient struct {
	bitwarden.PasswordManager
	err error
}

func (c *ephemeralSSHKeyErrorClient) GetItem(context.Context, models.Item) (*models.Item, error) {
	return nil, c.err
}

func (c *ephemeralSSHKeyErrorClient) FindItem(context.Context, ...bitwarden.ListObjectsOption) (*models.Item, error) {
	return nil, c.err
}

type ephemeralItemSSHKeyPersistenceCheck struct{}

func (ephemeralItemSSHKeyPersistenceCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	resp.Error = checkEphemeralItemSSHKeyPersistence(req.Plan)
}

func (ephemeralItemSSHKeyPersistenceCheck) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	resp.Error = checkEphemeralItemSSHKeyPersistence(req.State)
}

func checkEphemeralItemSSHKeyPersistence(value any) error {
	content, err := json.Marshal(value)
	if err != nil {
		return err
	}
	for _, marker := range []string{
		ephemeralSSHCredential,
		"ephemeral-public-key-must-not-be-persisted",
		"ephemeral-fingerprint-must-not-be-persisted",
		"ephemeral-ssh-notes-must-not-be-persisted",
	} {
		if strings.Contains(string(content), marker) {
			return fmt.Errorf("ephemeral SSH key item data was persisted")
		}
	}
	return nil
}
