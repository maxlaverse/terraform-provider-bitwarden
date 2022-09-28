package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

const (
	defaultBitwardenServerURL = "https://vault.bitwarden.com"
)

type LoginMethod int

const (
	LoginMethodPersonalAPIKey     LoginMethod = iota
	LoginMethodPassword           LoginMethod = iota
	LoginMethodProvidedSessionKey LoginMethod = iota
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				attributeSessionKey: {
					Type:        schema.TypeString,
					Description: descriptionSessionKey,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_SESSION", nil),
				},
				attributeMasterPassword: {
					Type:        schema.TypeString,
					Description: descriptionMasterPassword,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_PASSWORD", nil),
				},
				attributeClientID: {
					Type:         schema.TypeString,
					Description:  descriptionClientID,
					Optional:     true,
					RequiredWith: []string{attributeClientSecret},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTID", nil),
				},
				attributeClientSecret: {
					Type:         schema.TypeString,
					Description:  descriptionClientSecret,
					Optional:     true,
					RequiredWith: []string{attributeClientID},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTSECRET", nil),
				},
				attributeServer: {
					Type:        schema.TypeString,
					Description: descriptionServer,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_URL", defaultBitwardenServerURL),
				},
				attributeEmail: {
					Type:        schema.TypeString,
					Description: descriptionEmail,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_EMAIL", nil),
				},
				attributeVaultPath: {
					Type:        schema.TypeString,
					Description: descriptionVaultPath,
					Optional:    true,
					Default:     ".bitwarden/",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"bitwarden_item_login":       dataSourceItemLogin(),
				"bitwarden_item_secure_note": dataSourceItemSecureNote(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"bitwarden_folder":           resourceFolder(),
				"bitwarden_item_login":       resourceItemLogin(),
				"bitwarden_item_secure_note": resourceItemSecureNote(),
			},
		}

		p.ConfigureContextFunc = providerConfigure(version, p)
		return p
	}
}

func providerConfigure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

		bwClient, err := newBitwardenClient(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		sessionKey, hasSessionKey := d.GetOk(attributeSessionKey)
		if hasSessionKey {
			bwClient.SetSessionKey(sessionKey.(string))
		}

		err = ensureLoggedIn(d, bwClient)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return bwClient, nil
	}
}

func ensureLoggedIn(d *schema.ResourceData, bwClient bw.Client) error {
	status, err := logoutIfIdentityChanged(d, bwClient)
	if err != nil {
		return err
	}

	// Situation 1: the Vault is *unlocked* meaning we have a valid session key.
	if status.Status == bw.StatusUnlocked {
		return bwClient.Sync()
	}

	// Situation 2: the Vault is *locked* which means the Vault is already present locally
	//              and is ours.
	masterPassword := d.Get(attributeMasterPassword)
	if status.Status == bw.StatusLocked {
		err = bwClient.Unlock(masterPassword.(string))
		if err != nil {
			return err
		}

		return bwClient.Sync()
	}

	if status.Status != bw.StatusUnauthenticated {
		return fmt.Errorf("INTERNAL BUG: unsupported status: '%s'", status.Status)
	}

	// Situation 3: the Vault is *unauthenticated*
	loginMethod := loginMethod(d)
	switch loginMethod {
	case LoginMethodPersonalAPIKey:
		clientID := d.Get(attributeClientID)
		clientSecret := d.Get(attributeClientSecret)
		return bwClient.LoginWithAPIKey(masterPassword.(string), clientID.(string), clientSecret.(string))
	case LoginMethodPassword:
		email := d.Get(attributeEmail)
		return bwClient.LoginWithPassword(email.(string), masterPassword.(string))
	default:
		return fmt.Errorf("INTERNAL BUG: unsupported login method: %d", loginMethod)
	}
}

func loginMethod(d *schema.ResourceData) LoginMethod {
	_, hasClientID := d.GetOk(attributeClientID)
	_, hasClientSecret := d.GetOk(attributeClientSecret)

	if hasClientID && hasClientSecret {
		return LoginMethodPersonalAPIKey
	} else {
		return LoginMethodPassword
	}
}

func logoutIfIdentityChanged(d *schema.ResourceData, bwClient bw.Client) (*bw.Status, error) {
	status, err := bwClient.Status()
	if err != nil {
		return nil, err
	}

	email := d.Get(attributeEmail)
	serverURL := d.Get(attributeServer)
	if status.VaultOf(email.(string), serverURL.(string)) || status.FreshDataFile() {
		return status, nil
	}

	// We're not authenticated or authenticated against a different server.
	if status.Status != bw.StatusUnauthenticated {
		err = bwClient.Logout()
		if err != nil {
			return nil, err
		}
		status.Status = bw.StatusUnauthenticated
	}

	if status.ServerURL != serverURL.(string) {
		err = bwClient.SetServer(serverURL.(string))
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}

func newBitwardenClient(d *schema.ResourceData) (bw.Client, error) {
	opts := []bw.Options{}
	if ded, exists := d.GetOk(attributeVaultPath); exists {
		abs, err := filepath.Abs(ded.(string))
		if err != nil {
			return nil, err
		}
		opts = append(opts, bw.WithAppDataDir(abs))
	}
	if len(os.Getenv("DEBUG_BITWARDEN_DISABLE_SYNC")) > 0 {
		opts = append(opts, bw.DisableSync())
	}
	bwExecutable, err := exec.LookPath("bw")
	if err != nil {
		return nil, err
	}

	return bw.NewClient(bwExecutable, opts...), nil
}
