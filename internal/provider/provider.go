package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
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
					DefaultFunc: schema.EnvDefaultFunc("BW_URL", bitwarden.DefaultBitwardenServerURL),
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
					DefaultFunc: schema.EnvDefaultFunc("BITWARDENCLI_APPDATA_DIR", ".bitwarden/"),
				},
				attributeExtraCACertsPath: {
					Type:        schema.TypeString,
					Description: descriptionExtraCACertsPath,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("NODE_EXTRA_CA_CERTS", nil),
				},

				// Experimental
				attributeExperimental: {
					Description: descriptionExperimental,
					Type:        schema.TypeSet,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							attributeExperimentalEmbeddedClient: {
								Description: descriptionExperimentalEmbeddedClient,
								Type:        schema.TypeBool,
								Optional:    true,
							},
						},
					},
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"bitwarden_attachment":       dataSourceAttachment(),
				"bitwarden_folder":           dataSourceFolder(),
				"bitwarden_item_login":       dataSourceItemLogin(),
				"bitwarden_item_secure_note": dataSourceItemSecureNote(),
				"bitwarden_org_collection":   dataSourceOrgCollection(),
				"bitwarden_organization":     dataSourceOrganization(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"bitwarden_attachment":       resourceAttachment(),
				"bitwarden_folder":           resourceFolder(),
				"bitwarden_item_login":       resourceItemLogin(),
				"bitwarden_item_secure_note": resourceItemSecureNote(),
				"bitwarden_org_collection":   resourceOrgCollection(),
			},
		}

		p.ConfigureContextFunc = providerConfigure(version, p)
		return p
	}
}

func providerConfigure(version string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

		useEmbeddedClient := useExperimentalEmbeddedClient(d)

		if _, hasSessionKey := d.GetOk(attributeSessionKey); useEmbeddedClient && hasSessionKey {
			return nil, diag.Errorf("session key is not supported with the embedded client")
		}

		if useEmbeddedClient {
			bwClient, err := newEmbeddedPasswordManagerClient(ctx, d, version)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			err = ensureLoggedInEmbeddedPasswordManager(ctx, d, bwClient)
			if err != nil {
				return nil, diag.FromErr(err)
			}
			return bwClient, nil
		}

		bwClient, err := newCLIPasswordManagerClient(d, version)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		sessionKey, hasSessionKey := d.GetOk(attributeSessionKey)
		if hasSessionKey {
			bwClient.SetSessionKey(sessionKey.(string))
		}

		err = ensureLoggedInCLIPasswordManager(ctx, d, bwClient)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return bwClient, nil
	}
}

func useExperimentalEmbeddedClient(d *schema.ResourceData) bool {
	experimentalFeatures, hasExperimentalFeatures := d.GetOk(attributeExperimental)
	if !hasExperimentalFeatures {
		return false
	}
	if hasExperimentalFeatures {
		if experimentalFeatures.(*schema.Set).Len() == 0 {
			return false
		}
		embeddedClient, hasEmbeddedClient := experimentalFeatures.(*schema.Set).List()[0].(map[string]interface{})[attributeExperimentalEmbeddedClient]
		if hasEmbeddedClient && embeddedClient.(bool) {
			return true
		}
	}
	return false
}

func ensureLoggedInCLIPasswordManager(ctx context.Context, d *schema.ResourceData, bwClient bwcli.PasswordManagerClient) error {
	status, err := bwClient.Status(ctx)
	if err != nil {
		return err
	}

	err = logoutIfIdentityChanged(ctx, d, bwClient, status)
	if err != nil {
		return err
	}

	// Scenario 1: The Vault is already *unlocked*, there is nothing else to
	//             be done. This should happen when a session key is provided.
	//             => return
	if status.Status == bwcli.StatusUnlocked {
		return bwClient.Sync(ctx)
	}

	// Scenario 2: The Vault is *locked* and we have a master password. This
	//             happens when the Vault is already cached locally.
	//             => unlock and return
	masterPassword, hasMasterPassword := d.GetOk(attributeMasterPassword)
	if hasMasterPassword && status.Status == bwcli.StatusLocked {
		err = bwClient.Unlock(ctx, masterPassword.(string))
		if err != nil {
			return err
		}

		return bwClient.Sync(ctx)
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
		return bwClient.LoginWithAPIKey(ctx, masterPassword.(string), clientID.(string), clientSecret.(string))
	case LoginMethodPassword:
		email := d.Get(attributeEmail)
		return bwClient.LoginWithPassword(ctx, email.(string), masterPassword.(string))
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

func logoutIfIdentityChanged(ctx context.Context, d *schema.ResourceData, bwClient bwcli.PasswordManagerClient, status *bwcli.Status) error {
	email := d.Get(attributeEmail).(string)
	serverURL := d.Get(attributeServer).(string)

	if (status.Status == bwcli.StatusLocked || status.Status == bwcli.StatusUnlocked) && (!status.VaultOfUser(email) || !status.VaultFromServer(serverURL)) {
		status.Status = bwcli.StatusUnauthenticated

		tflog.Warn(ctx, "Logging out as the local Vault belongs to a different identity", map[string]interface{}{"vault_email": status.UserEmail, "vault_server": status.ServerURL, "provider_email": email, "provider_server": serverURL})
		err := bwClient.Logout(ctx)
		if err != nil {
			return err
		}
	}

	if !status.VaultFromServer(serverURL) {
		err := bwClient.SetServer(ctx, serverURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func newCLIPasswordManagerClient(d *schema.ResourceData, version string) (bwcli.PasswordManagerClient, error) {
	opts := []bwcli.Options{}
	if vaultPath, exists := d.GetOk(attributeVaultPath); exists {
		abs, err := filepath.Abs(vaultPath.(string))
		if err != nil {
			return nil, err
		}
		opts = append(opts, bwcli.WithAppDataDir(abs))
	}

	if extraCACertsPath, exists := d.GetOk(attributeExtraCACertsPath); exists {
		opts = append(opts, bwcli.WithExtraCACertsPath(extraCACertsPath.(string)))
	}

	if version == versionDev {
		// During development, we disable retry backoffs to make some operations faster.
		opts = append(opts, bwcli.DisableRetryBackoff())
	}

	bwExecutable, err := exec.LookPath("bw")
	if err != nil {
		return nil, err
	}

	return bwcli.NewPasswordManagerClient(bwExecutable, opts...), nil
}

func newEmbeddedPasswordManagerClient(ctx context.Context, d *schema.ResourceData, version string) (bitwarden.PasswordManager, error) {
	deviceId, err := getOrGenerateDeviceIdentifier(ctx)
	if err != nil {
		return nil, err
	}

	webapiOpts := []webapi.Options{}
	if version == versionDev {
		// During development, we don't want to wait on any sporadic errors.
		webapiOpts = append(webapiOpts, webapi.DisableRetries())
	}

	serverURL := d.Get(attributeServer).(string)
	return embedded.NewPasswordManagerClient(serverURL, deviceId, embedded.WithPasswordManagerHttpOptions(webapiOpts...)), nil
}

func getOrGenerateDeviceIdentifier(ctx context.Context) (string, error) {
	deviceIdBytes, err := os.ReadFile(".bitwarden/device_identifier")
	if err == nil {
		deviceId := string(deviceIdBytes)
		tflog.Info(ctx, "Read device identifier from disk", map[string]interface{}{"device_id": deviceId})
		return strings.TrimSpace(deviceId), nil
	}

	deviceId := embedded.NewDeviceIdentifier()
	err = os.Mkdir(".bitwarden", 0700)
	if err != nil && !os.IsExist(err) {
		tflog.Error(ctx, "Failed to create .bitwarden directory", map[string]interface{}{"error": err})
		return "", err
	}
	err = os.WriteFile(".bitwarden/device_identifier", []byte(deviceId), 0600)
	if err != nil {
		tflog.Error(ctx, "Failed to store device identifier", map[string]interface{}{"error": err})
		return "", err
	}

	tflog.Info(ctx, "Generated device identifier", map[string]interface{}{"device_id": deviceId})
	return deviceId, nil
}

func ensureLoggedInEmbeddedPasswordManager(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) error {
	masterPassword, hasMasterPassword := d.GetOk(attributeMasterPassword)
	if !hasMasterPassword {
		return fmt.Errorf("master password is required")
	}

	loginMethod := loginMethod(d)
	switch loginMethod {
	case LoginMethodPersonalAPIKey:
		clientID := d.Get(attributeClientID)
		clientSecret := d.Get(attributeClientSecret)
		return bwClient.LoginWithAPIKey(ctx, masterPassword.(string), clientID.(string), clientSecret.(string))
	case LoginMethodPassword:
		email := d.Get(attributeEmail)
		return bwClient.LoginWithPassword(ctx, email.(string), masterPassword.(string))
	}

	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: 'BUG')")
}
