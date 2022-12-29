package provider

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bw"
)

type LoginMethod int

const (
	LoginMethodPersonalAPIKey LoginMethod = iota
	LoginMethodPassword       LoginMethod = iota
	LoginMethodNone           LoginMethod = iota
)

const (
	versionDev = "dev"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				// Attributes which depend on each other
				attributeMasterPassword: {
					Type:          schema.TypeString,
					Description:   descriptionMasterPassword,
					ConflictsWith: []string{attributeSessionKey},
					AtLeastOneOf:  []string{attributeSessionKey},
					Optional:      true,
					DefaultFunc:   schema.EnvDefaultFunc("BW_PASSWORD", nil),
				},
				attributeSessionKey: {
					Type:          schema.TypeString,
					Description:   descriptionSessionKey,
					ConflictsWith: []string{attributeMasterPassword},
					AtLeastOneOf:  []string{attributeMasterPassword},
					Optional:      true,
					DefaultFunc:   schema.EnvDefaultFunc("BW_SESSION", nil),
				},
				attributeClientID: {
					Type:         schema.TypeString,
					Description:  descriptionClientID,
					Optional:     true,
					RequiredWith: []string{attributeClientSecret, attributeMasterPassword},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTID", nil),
				},
				attributeClientSecret: {
					Type:         schema.TypeString,
					Description:  descriptionClientSecret,
					Optional:     true,
					RequiredWith: []string{attributeClientID, attributeMasterPassword},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTSECRET", nil),
				},

				// Standalone attributes
				attributeServer: {
					Type:        schema.TypeString,
					Description: descriptionServer,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_URL", bw.DefaultBitwardenServerURL),
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
				"bitwarden_item_attachment":  resourceItemAttachment(),
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

		bwClient, err := newBitwardenClient(d, version)
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
	status, err := bwClient.Status()
	if err != nil {
		return err
	}

	err = logoutIfIdentityChanged(d, bwClient, status)
	if err != nil {
		return err
	}

	// Scenario 1: The Vault is already *unlocked*, there is nothing else to
	//             be done. This should happen when a session key is provided.
	//             => return
	if status.Status == bw.StatusUnlocked {
		return bwClient.Sync()
	}

	// Scenario 2: The Vault is *locked* and we have a master password. This
	//             happens when the Vault is already cached locally.
	//             => unlock and return
	masterPassword, hasMasterPassword := d.GetOk(attributeMasterPassword)
	if hasMasterPassword && status.Status == bw.StatusLocked {
		err = bwClient.Unlock(masterPassword.(string))
		if err != nil {
			return err
		}

		return bwClient.Sync()
	}

	// Scenario 3: We need to login and have enough information to do so.
	//             Happens if the Vault is not present locally, or it doesn't
	//             belong to us.
	//             => login and return
	//
	// Note: We don't trigger a manual 'sync' as login operations already do.
	loginMethod := loginMethod(d)
	switch loginMethod {
	case LoginMethodPersonalAPIKey:
		clientID := d.Get(attributeClientID)
		clientSecret := d.Get(attributeClientSecret)
		return bwClient.LoginWithAPIKey(masterPassword.(string), clientID.(string), clientSecret.(string))
	case LoginMethodPassword:
		email := d.Get(attributeEmail)
		return bwClient.LoginWithPassword(email.(string), masterPassword.(string))
	}

	// Scenario 4: We need to login but don't have the information to do so.
	//             This is a situation we can't get out from.
	//             => failure
	if _, hasSessionKey := d.GetOk(attributeSessionKey); hasSessionKey {
		return fmt.Errorf("unable to unlock Vault with provided session key (status: %s)", status.Status)
	}

	// We should have caught already scenarios up to this point. If we haven't, it means this method's
	// implementation is wrong or the provider parameters are.
	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: '%s')", status.Status)
}

func loginMethod(d *schema.ResourceData) LoginMethod {
	_, hasClientID := d.GetOk(attributeClientID)
	_, hasClientSecret := d.GetOk(attributeClientSecret)
	_, hasMasterPassword := d.GetOk(attributeMasterPassword)

	if hasClientID && hasClientSecret {
		return LoginMethodPersonalAPIKey
	} else if hasMasterPassword {
		return LoginMethodPassword
	}

	return LoginMethodNone
}

func logoutIfIdentityChanged(d *schema.ResourceData, bwClient bw.Client, status *bw.Status) error {
	email := d.Get(attributeEmail).(string)
	serverURL := d.Get(attributeServer).(string)

	if (status.Status == bw.StatusLocked || status.Status == bw.StatusUnlocked) && (!status.VaultOfUser(email) || !status.VaultFromServer(serverURL)) {
		status.Status = bw.StatusUnauthenticated

		log.Printf("Logging out as the local Vault belongs to a different identity (vault: '%v' on  '%s', provider: '%v' on '%s')\n", status.UserEmail, status.ServerURL, email, status.ServerURL)
		err := bwClient.Logout()
		if err != nil {
			return err
		}
	}

	if !status.VaultFromServer(serverURL) {
		err := bwClient.SetServer(serverURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func newBitwardenClient(d *schema.ResourceData, version string) (bw.Client, error) {
	opts := []bw.Options{}
	if ded, exists := d.GetOk(attributeVaultPath); exists {
		abs, err := filepath.Abs(ded.(string))
		if err != nil {
			return nil, err
		}
		opts = append(opts, bw.WithAppDataDir(abs))
	}
	if version == versionDev {
		// During development, we disable Vault synchronization to make some operations faster.
		opts = append(opts, bw.DisableSync())
	}
	bwExecutable, err := exec.LookPath("bw")
	if err != nil {
		return nil, err
	}

	return bw.NewClient(bwExecutable, opts...), nil
}
