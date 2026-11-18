//go:build offline

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func TestItemLoginWriteOnlyConfigValidation(t *testing.T) {
	for _, tt := range []struct {
		name, body, want string
	}{
		{"password_wo needs version", `password_wo = "s"`, "password_wo_version"},
		{"version needs password_wo", `password_wo_version = 1`, "password_wo"},
		{"password_wo conflicts with password", `
			password = "a"
			password_wo = "s"
			password_wo_version = 1`, "password"},
		{"empty password_wo", `
			password_wo = ""
			password_wo_version = 1`, "at least 1"},
		{"negative version", `
			password_wo = "s"
			password_wo_version = -1`, "at least 0"},
		{"fields_wo needs version", `fields_wo = { a = "s" }`, "fields_wo_version"},
		{"version needs fields_wo", `fields_wo_version = 1`, "fields_wo"},
		{"empty fields_wo", `
			fields_wo = {}
			fields_wo_version = 1`, "at least 1"},
		{"empty fields_wo key", `
			fields_wo = { "" = "s" }
			fields_wo_version = 1`, "at least 1"},
		{"field block collides with fields_wo", `
			field {
				name = "dup"
				text = "x"
			}
			fields_wo = { dup = "s" }
			fields_wo_version = 1`, "Duplicate custom field"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: ephemeralCredentialFactories(nil, models.Secret{}),
				Steps: []resource.TestStep{{
					Config: `
provider "bitwarden" {
  access_token = "unused"
}

resource "bitwarden_item_login" "probe" {
  name = "probe"
  ` + tt.body + `
}
`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tt.want),
				}},
			})
		})
	}
}

// fakeItemStore is an in-memory PasswordManager that, like the real API, replaces an item wholesale on edit.
type fakeItemStore struct {
	bitwarden.PasswordManager
	items map[string]models.Item
}

func (f *fakeItemStore) CreateItem(_ context.Context, obj models.Item) (*models.Item, error) {
	obj.ID = fmt.Sprintf("item-%d", len(f.items)+1)
	f.items[obj.ID] = obj
	return &obj, nil
}

func (f *fakeItemStore) EditItem(_ context.Context, obj models.Item) (*models.Item, error) {
	f.items[obj.ID] = obj
	return &obj, nil
}

func (f *fakeItemStore) GetItem(_ context.Context, obj models.Item) (*models.Item, error) {
	item, ok := f.items[obj.ID]
	if !ok {
		return nil, models.ErrObjectNotFound
	}
	return &item, nil
}

func (f *fakeItemStore) DeleteItem(_ context.Context, obj models.Item) error {
	delete(f.items, obj.ID)
	return nil
}

func (f *fakeItemStore) check(want func(models.Item) error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if len(f.items) != 1 {
			return fmt.Errorf("want 1 stored item, got %d", len(f.items))
		}
		for _, it := range f.items {
			return want(it)
		}
		return nil
	}
}

func TestItemLoginWriteOnlyUpdates(t *testing.T) {
	store := &fakeItemStore{items: map[string]models.Item{}}
	cfg := func(name, pw string, pwVersion int, field string, fieldVersion int) string {
		return fmt.Sprintf(`
provider "bitwarden" {
  access_token = "bootstrap"
}

resource "bitwarden_item_login" "probe" {
  name                = %q
  password_wo         = %q
  password_wo_version = %d
  field {
    name = "kind"
    text = "generic"
  }
  fields_wo         = { secret = %q }
  fields_wo_version = %d
}
`, name, pw, pwVersion, field, fieldVersion)
	}
	stored := func(password, secret string) resource.TestCheckFunc {
		return store.check(func(it models.Item) error {
			fields := map[string]string{}
			for _, f := range it.Fields {
				fields[f.Name] = f.Value
			}
			if it.Login.Password != password || fields["secret"] != secret || fields["kind"] != "generic" {
				return fmt.Errorf("stored password %q fields %v, want password %q secret %q", it.Login.Password, fields, password, secret)
			}
			return nil
		})
	}
	noSecretInState := resource.ComposeTestCheckFunc(
		resource.TestCheckNoResourceAttr("bitwarden_item_login.probe", "password"),
		resource.TestCheckNoResourceAttr("bitwarden_item_login.probe", "password_wo"),
		resource.TestCheckNoResourceAttr("bitwarden_item_login.probe", "fields_wo"),
		resource.TestCheckResourceAttr("bitwarden_item_login.probe", "field.#", "1"),
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ephemeralCredentialFactories(&ProviderClients{PasswordManager: store}, models.Secret{}),
		Steps: []resource.TestStep{
			{Config: cfg("probe", "pw-1", 1, "sec-1", 1), Check: resource.ComposeTestCheckFunc(noSecretInState, stored("pw-1", "sec-1"))},
			// unrelated edit, versions unchanged: neither rotated nor dropped
			{Config: cfg("renamed", "pw-ignored", 1, "sec-ignored", 1), Check: resource.ComposeTestCheckFunc(noSecretInState, stored("pw-1", "sec-1"))},
			// bumping only the password version leaves the write-only field alone
			{Config: cfg("renamed", "pw-2", 2, "sec-ignored", 1), Check: resource.ComposeTestCheckFunc(noSecretInState, stored("pw-2", "sec-1"))},
			{Config: cfg("renamed", "pw-2", 2, "sec-2", 2), Check: resource.ComposeTestCheckFunc(noSecretInState, stored("pw-2", "sec-2"))},
		},
	})
}
