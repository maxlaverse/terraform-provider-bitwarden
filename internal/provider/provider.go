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

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				attributeMasterPassword: {
					Type:        schema.TypeString,
					Description: descriptionMasterPassword,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_PASSWORD", ""),
				},
				attributeClientID: {
					Type:        schema.TypeString,
					Description: descriptionClientID,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_CLIENTID", ""),
				},
				attributeClientSecret: {
					Type:        schema.TypeString,
					Description: descriptionClientSecret,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_CLIENTSECRET", ""),
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
					DefaultFunc: schema.EnvDefaultFunc("BW_EMAIL", ""),
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

		err = ensureLoggedIn(d, bwClient)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return bwClient, nil
	}
}

func ensureLoggedIn(d *schema.ResourceData, bwClient bw.Client) error {
	masterPassword := d.Get(attributeMasterPassword)
	serverURL := d.Get(attributeServer)
	email := d.Get(attributeEmail)

	status, err := bwClient.Status()
	if err != nil {
		return err
	}

	if status.ServerURL != serverURL.(string) || status.UserEmail != email.(string) {
		if status.ServerURL != serverURL.(string) {
			err = bwClient.SetServer(serverURL.(string))
			if err != nil {
				return err
			}
		}

		if status.UserEmail != email.(string) {
			if status.Status != bw.StatusUnauthenticated {
				// We're already logged in as a different user, log out
				err = bwClient.Logout()
				if err != nil {
					return err
				}
			}

			clientID, hasClientID := d.GetOk(attributeClientID)
			clientSecret, hasClientSecret := d.GetOk(attributeClientSecret)

			if hasClientID && hasClientSecret {
				err = bwClient.LoginWithAPIKeyAndUnlock(masterPassword.(string), clientID.(string), clientSecret.(string))
				if err != nil {
					return err
				}
			} else if !hasClientID && !hasClientSecret {
				err = bwClient.LoginWithPassword(email.(string), masterPassword.(string))
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("%s and %s must be set together", attributeClientID, attributeClientSecret)
			}
		}
	}

	if !bwClient.HasSessionKey() {
		err = bwClient.Unlock(masterPassword.(string))
		if err != nil {
			return err
		}

		err = bwClient.Sync()
		if err != nil {
			return err
		}
	}

	return nil
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
