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

const ephemeralLoginCredential = "ephemeral-password-must-not-be-persisted"

func TestEphemeralItemLoginProviderConfiguration(t *testing.T) {
	item := ephemeralLoginItem()
	for _, tt := range []struct {
		name     string
		selector string
	}{
		{name: "by ID", selector: `id = "item-id"`},
		{name: "by search", selector: `search = "VPN/Credentials"`},
		{name: "by search and URL", selector: "search = \"VPN/Credentials\"\n  filter_url = \"https://vpn.example.com\""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			check := ephemeralItemLoginPersistenceCheck{}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralLoginFactories(nil, item),
				Steps: []resource.TestStep{{
					Config:            ephemeralItemLoginConsumerConfig(tt.selector),
					ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{check}},
					ConfigStateChecks: []statecheck.StateCheck{check},
					Check:             resource.TestCheckOutput("folder_name", "public-folder"),
				}},
			})
		})
	}
}

func TestEphemeralItemLoginErrors(t *testing.T) {
	item := ephemeralLoginItem()
	for _, tt := range []struct {
		name     string
		clients  *ProviderClients
		selector string
		want     string
	}{
		{name: "no Password Manager client", clients: &ProviderClients{}, selector: `id = "item-id"`, want: "Provider not configured for Password Manager"},
		{name: "server error details", clients: &ProviderClients{PasswordManager: &ephemeralLoginErrorClient{err: &webapi.HTTPError{StatusCode: 503, Message: "Service temporarily unavailable"}}}, selector: `search = "VPN/Credentials"`, want: "Service temporarily unavailable"},
		{name: "object not found", clients: &ProviderClients{PasswordManager: &ephemeralLoginClient{item: item}}, selector: `id = "missing-item"`, want: "object not found"},
		{name: "no object found matching filter", clients: &ProviderClients{PasswordManager: &ephemeralLoginClient{item: item}}, selector: `search = "missing-item"`, want: "no object found matching the filter"},
		{name: "url filter mismatch", clients: &ProviderClients{PasswordManager: &ephemeralLoginClient{item: item}}, selector: "search = \"VPN/Credentials\"\n  filter_url = \"https://missing.example.com\"", want: "no object found matching the filter"},
		{name: "too many objects found", clients: &ProviderClients{PasswordManager: &ephemeralLoginErrorClient{err: models.ErrTooManyObjectsFound}}, selector: `search = "VPN/Credentials"`, want: "too many objects found"},
		{name: "item type mismatch", clients: &ProviderClients{PasswordManager: &ephemeralLoginErrorClient{err: models.ErrItemTypeMismatch}}, selector: `id = "item-id"`, want: "returned object type does not match requested object type"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralLoginFactories(tt.clients, item),
				Steps: []resource.TestStep{{
					Config:      ephemeralItemLoginConsumerConfig(tt.selector),
					ExpectError: regexp.MustCompile(tt.want),
				}},
			})
		})
	}
}

func ephemeralLoginItem() models.Item {
	return models.Item{
		ID:     "item-id",
		Name:   "VPN/Credentials",
		Object: models.ObjectTypeItem,
		Type:   models.ItemTypeLogin,
		Notes:  "ephemeral-notes-must-not-be-persisted",
		Login: models.Login{
			Username: "ephemeral-username-must-not-be-persisted",
			Password: ephemeralLoginCredential,
			Totp:     "ephemeral-totp-must-not-be-persisted",
			URIs:     []models.LoginURI{{URI: "https://vpn.example.com"}},
		},
	}
}

func ephemeralItemLoginConsumerConfig(selector string) string {
	return fmt.Sprintf(`
provider "bitwarden" {
  alias       = "source"
  email       = "source@example.com"
  session_key = "bootstrap"
}

ephemeral "bitwarden_item_login" "credential" {
  provider = bitwarden.source
  %s
}

provider "bitwarden" {
  alias       = "consumer"
  email       = "consumer@example.com"
  session_key = ephemeral.bitwarden_item_login.credential.password
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

func ephemeralLoginFactories(clients *ProviderClients, item models.Item) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bitwarden": func() (tfprotov6.ProviderServer, error) {
			return providerserver.NewProtocol6(&ephemeralLoginProvider{
				bitwardenProvider: &bitwardenProvider{},
				sourceClients:     clients,
				item:              item,
				credential:        item.Login.Password,
			})(), nil
		},
	}
}

// Use the real schemas and resource implementations, replacing only backend
// access so the CLI can exercise ephemeral provider configuration offline.
type ephemeralLoginProvider struct {
	*bitwardenProvider
	sourceClients *ProviderClients
	item          models.Item
	credential    string
}

func (p *ephemeralLoginProvider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
	var model bitwardenProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	clients := &ProviderClients{PasswordManager: &ephemeralLoginClient{
		sessionKey: model.SessionKey.ValueString(), item: p.item, credential: p.credential,
	}}
	if model.SessionKey.ValueString() == "bootstrap" && p.sourceClients != nil {
		clients = p.sourceClients
	}
	resp.ResourceData = clients
	resp.DataSourceData = clients
	resp.EphemeralResourceData = clients
}

type ephemeralLoginClient struct {
	bitwarden.PasswordManager
	sessionKey string
	item       models.Item
	credential string
}

func (c *ephemeralLoginClient) GetItem(_ context.Context, item models.Item) (*models.Item, error) {
	if item.ID != c.item.ID {
		return nil, models.ErrObjectNotFound
	}
	if item.Type != 0 && item.Type != c.item.Type {
		return nil, models.ErrItemTypeMismatch
	}
	return &c.item, nil
}

func (c *ephemeralLoginClient) FindItem(_ context.Context, options ...bitwarden.ListObjectsOption) (*models.Item, error) {
	filter := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if filter.SearchFilter != "" && filter.SearchFilter != c.item.Name && filter.SearchFilter != c.item.Login.Username {
		return nil, models.ErrNoObjectFoundMatchingFilter
	}
	if filter.UrlFilter != "" {
		matched := false
		for _, uri := range c.item.Login.URIs {
			if uri.URI == filter.UrlFilter {
				matched = true
				break
			}
		}
		if !matched {
			return nil, models.ErrNoObjectFoundMatchingFilter
		}
	}
	if filter.ItemType != 0 && filter.ItemType != c.item.Type {
		return nil, models.ErrNoObjectFoundMatchingFilter
	}
	return &c.item, nil
}

func (c *ephemeralLoginClient) GetFolder(_ context.Context, folder models.Folder) (*models.Folder, error) {
	if c.sessionKey != c.credential {
		return nil, fmt.Errorf("consumer did not receive the ephemeral credential")
	}
	return &models.Folder{ID: folder.ID, Name: "public-folder"}, nil
}

type ephemeralLoginErrorClient struct {
	bitwarden.PasswordManager
	err error
}

func (c *ephemeralLoginErrorClient) GetItem(context.Context, models.Item) (*models.Item, error) {
	return nil, c.err
}

func (c *ephemeralLoginErrorClient) FindItem(context.Context, ...bitwarden.ListObjectsOption) (*models.Item, error) {
	return nil, c.err
}

type ephemeralItemLoginPersistenceCheck struct{}

func (ephemeralItemLoginPersistenceCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	resp.Error = checkEphemeralItemLoginPersistence(req.Plan)
}

func (ephemeralItemLoginPersistenceCheck) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	resp.Error = checkEphemeralItemLoginPersistence(req.State)
}

func checkEphemeralItemLoginPersistence(value any) error {
	content, err := json.Marshal(value)
	if err != nil {
		return err
	}
	for _, marker := range []string{
		ephemeralLoginCredential,
		"ephemeral-username-must-not-be-persisted",
		"ephemeral-totp-must-not-be-persisted",
		"ephemeral-notes-must-not-be-persisted",
	} {
		if strings.Contains(string(content), marker) {
			return fmt.Errorf("ephemeral login item data was persisted")
		}
	}
	return nil
}
