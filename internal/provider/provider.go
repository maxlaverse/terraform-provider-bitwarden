package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwscli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

type LoginMethod int

const (
	LoginMethodPersonalAPIKey LoginMethod = iota
	LoginMethodPassword       LoginMethod = iota
	LoginMethodNone           LoginMethod = iota
	LoginMethodAccessToken    LoginMethod = iota
)

const (
	versionTestDefault         = ""
	versionTestDisabledRetries = "--disable-retries--"
	versionTestSkippedLogin    = "--skip-login--"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				// Attributes which depend on each other
				schema_definition.AttributeMasterPassword: {
					Type:          schema.TypeString,
					Description:   schema_definition.DescriptionMasterPassword,
					ConflictsWith: []string{schema_definition.AttributeSessionKey, schema_definition.AttributeAccessToken},
					AtLeastOneOf:  []string{schema_definition.AttributeSessionKey, schema_definition.AttributeAccessToken},
					Optional:      true,
					DefaultFunc:   schema.EnvDefaultFunc("BW_PASSWORD", nil),
				},
				schema_definition.AttributeSessionKey: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionSessionKey,
					AtLeastOneOf: []string{schema_definition.AttributeMasterPassword, schema_definition.AttributeAccessToken},
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("BW_SESSION", nil),
				},
				schema_definition.AttributeClientID: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionClientID,
					Optional:     true,
					RequiredWith: []string{schema_definition.AttributeClientSecret, schema_definition.AttributeMasterPassword},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTID", nil),
				},
				schema_definition.AttributeClientSecret: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionClientSecret,
					Optional:     true,
					RequiredWith: []string{schema_definition.AttributeClientID, schema_definition.AttributeMasterPassword},
					DefaultFunc:  schema.EnvDefaultFunc("BW_CLIENTSECRET", nil),
				},
				schema_definition.AttributeAccessToken: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionAccessToken,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BWS_ACCESS_TOKEN", nil),
				},

				// Standalone attributes
				schema_definition.AttributeServer: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionServer,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("BW_URL", bitwarden.DefaultBitwardenServerURL),
				},
				schema_definition.AttributeProviderEmail: {
					Type:         schema.TypeString,
					Description:  schema_definition.DescriptionProviderEmail,
					Optional:     true,
					AtLeastOneOf: []string{schema_definition.AttributeAccessToken, schema_definition.AttributeClientID, schema_definition.AttributeSessionKey},
					DefaultFunc:  schema.EnvDefaultFunc("BW_EMAIL", nil),
				},
				schema_definition.AttributeVaultPath: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionVaultPath,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("BITWARDENCLI_APPDATA_DIR", ".bitwarden/"),
				},
				schema_definition.AttributeExtraCACertsPath: {
					Type:        schema.TypeString,
					Description: schema_definition.DescriptionExtraCACertsPath,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("NODE_EXTRA_CA_CERTS", nil),
				},

				// Experimental
				schema_definition.AttributeExperimental: {
					Description: schema_definition.DescriptionExperimental,
					Type:        schema.TypeSet,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							schema_definition.AttributeExperimentalEmbeddedClient: {
								Description: schema_definition.DescriptionExperimentalEmbeddedClient,
								Type:        schema.TypeBool,
								Optional:    true,
							},
							schema_definition.AttributeExperimentalDisableSyncAfterWriteVerification: {
								Description: schema_definition.DescriptionExperimentalDisableSyncAfterWriteVerification,
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
				"bitwarden_item_ssh_key":     dataSourceItemSSHKey(),
				"bitwarden_org_collection":   dataSourceOrgCollection(),
				"bitwarden_org_group":        dataSourceOrgGroup(),
				"bitwarden_org_member":       dataSourceOrgMember(),
				"bitwarden_organization":     dataSourceOrganization(),
				"bitwarden_project":          dataSourceProject(),
				"bitwarden_secret":           dataSourceSecret(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"bitwarden_attachment":       resourceAttachment(),
				"bitwarden_folder":           resourceFolder(),
				"bitwarden_item_login":       resourceItemLogin(),
				"bitwarden_item_secure_note": resourceItemSecureNote(),
				"bitwarden_item_ssh_key":     resourceItemSSHKey(),
				"bitwarden_org_collection":   resourceOrgCollection(),
				"bitwarden_project":          resourceProject(),
				"bitwarden_secret":           resourceSecret(),
			},
		}

		p.ConfigureContextFunc = providerConfigure(version, p)
		return p
	}
}

func providerConfigure(version string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	shouldLogin := !strings.Contains(version, versionTestSkippedLogin)

	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

		_, hasAccessToken := d.GetOk(schema_definition.AttributeAccessToken)
		useEmbeddedClient := useExperimentalEmbeddedClient(d)

		if _, hasSessionKey := d.GetOk(schema_definition.AttributeSessionKey); useEmbeddedClient && hasSessionKey {
			return nil, diag.Errorf("session key is not supported with the embedded client")
		}

		if useEmbeddedClient && !hasAccessToken {
			bwClient, err := newEmbeddedPasswordManagerClient(ctx, d, version)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			if shouldLogin {
				err = ensureLoggedInEmbeddedPasswordManager(ctx, d, bwClient)
				if err != nil {
					return nil, diag.FromErr(err)
				}
			}
			return bwClient, nil
		} else if useEmbeddedClient && hasAccessToken {
			bwsClient, err := newEmbeddedSecretsManagerClient(ctx, d, version)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			if shouldLogin {
				err = ensureLoggedInEmbeddedSecretsManager(ctx, d, bwsClient)
				if err != nil {
					return nil, diag.FromErr(err)
				}
			}
			return bwsClient, nil
		} else if !useEmbeddedClient && hasAccessToken {
			bwsClient, err := newCLISecretsManagerClient(ctx, d, version)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			return bwsClient, nil
		}

		bwClient, err := newCLIPasswordManagerClient(d, version)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		sessionKey, hasSessionKey := d.GetOk(schema_definition.AttributeSessionKey)
		if hasSessionKey {
			bwClient.SetSessionKey(sessionKey.(string))
		}

		if shouldLogin {
			err = ensureLoggedInCLIPasswordManager(ctx, d, bwClient)
			if err != nil {
				return nil, diag.FromErr(err)
			}
		}

		return bwClient, nil
	}
}

func useExperimentalEmbeddedClient(d *schema.ResourceData) bool {
	return hasExperimentalFeature(d, schema_definition.AttributeExperimentalEmbeddedClient)
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
	masterPassword, hasMasterPassword := d.GetOk(schema_definition.AttributeMasterPassword)
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
		clientID := d.Get(schema_definition.AttributeClientID)
		clientSecret := d.Get(schema_definition.AttributeClientSecret)
		return bwClient.LoginWithAPIKey(ctx, masterPassword.(string), clientID.(string), clientSecret.(string))
	case LoginMethodPassword:
		email := d.Get(schema_definition.AttributeProviderEmail)
		return bwClient.LoginWithPassword(ctx, email.(string), masterPassword.(string))
	}

	// Scenario 4: We need to login but don't have the information to do so.
	//             This is a situation we can't get out from.
	//             => failure
	if _, hasSessionKey := d.GetOk(schema_definition.AttributeSessionKey); hasSessionKey {
		return fmt.Errorf("unable to unlock Vault with provided session key (status: %s)", status.Status)
	}

	// We should have caught already scenarios up to this point. If we haven't, it means this method's
	// implementation is wrong or the provider parameters are.
	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: '%s')", status.Status)
}

func loginMethod(d *schema.ResourceData) LoginMethod {
	_, hasClientID := d.GetOk(schema_definition.AttributeClientID)
	_, hasClientSecret := d.GetOk(schema_definition.AttributeClientSecret)
	_, hasAccessToken := d.GetOk(schema_definition.AttributeAccessToken)
	_, hasMasterPassword := d.GetOk(schema_definition.AttributeMasterPassword)

	if hasAccessToken {
		return LoginMethodAccessToken
	} else if hasClientID && hasClientSecret {
		return LoginMethodPersonalAPIKey
	} else if hasMasterPassword {
		return LoginMethodPassword
	}

	return LoginMethodNone
}

func logoutIfIdentityChanged(ctx context.Context, d *schema.ResourceData, bwClient bwcli.PasswordManagerClient, status *bwcli.Status) error {
	serverURL := d.Get(schema_definition.AttributeServer).(string)
	email, emailProvided := d.GetOk(schema_definition.AttributeProviderEmail)
	vaultBelongsToEmailAndServer := (!emailProvided || status.VaultOfUser(email.(string))) && status.VaultFromServer(serverURL)

	if (status.Status == bwcli.StatusLocked || status.Status == bwcli.StatusUnlocked) && !vaultBelongsToEmailAndServer {
		status.Status = bwcli.StatusUnauthenticated

		tflog.Warn(ctx, "Logging out as the local Vault belongs to a different identity", map[string]interface{}{"vault_email": status.UserEmail, "vault_server": status.ServerURL, "provider_server": serverURL})
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
	if vaultPath, exists := d.GetOk(schema_definition.AttributeVaultPath); exists {
		abs, err := filepath.Abs(vaultPath.(string))
		if err != nil {
			return nil, err
		}
		opts = append(opts, bwcli.WithAppDataDir(abs))
	}

	if extraCACertsPath, exists := d.GetOk(schema_definition.AttributeExtraCACertsPath); exists {
		opts = append(opts, bwcli.WithExtraCACertsPath(extraCACertsPath.(string)))
	}

	if version == versionTestDisabledRetries {
		// During development, we disable retry backoffs to make some operations faster.
		opts = append(opts, bwcli.DisableRetryBackoff())
	}

	return bwcli.NewPasswordManagerClient(opts...), nil
}

func newEmbeddedPasswordManagerClient(ctx context.Context, d *schema.ResourceData, version string) (bitwarden.PasswordManager, error) {
	deviceId, err := getOrGenerateDeviceIdentifier(ctx)
	if err != nil {
		return nil, err
	}

	opts := []embedded.PasswordManagerOptions{
		embedded.WithPasswordManagerHttpOptions(buildWebapiOptions(version)...),
	}

	if hasExperimentalFeature(d, schema_definition.AttributeExperimentalDisableSyncAfterWriteVerification) {
		opts = append(opts, embedded.DisableFailOnSyncAfterWriteVerification())
	}

	serverURL := d.Get(schema_definition.AttributeServer).(string)
	return embedded.NewPasswordManagerClient(serverURL, deviceId, version, opts...), nil
}

func newEmbeddedSecretsManagerClient(ctx context.Context, d *schema.ResourceData, version string) (bitwarden.SecretsManager, error) {
	deviceId, err := getOrGenerateDeviceIdentifier(ctx)
	if err != nil {
		return nil, err
	}

	serverURL := d.Get(schema_definition.AttributeServer).(string)
	return embedded.NewSecretsManagerClient(serverURL, deviceId, version, embedded.WithSecretsManagerHttpOptions(buildWebapiOptions(version)...)), nil
}

func newCLISecretsManagerClient(ctx context.Context, d *schema.ResourceData, version string) (bitwarden.SecretsManager, error) {
	serverURL := d.Get(schema_definition.AttributeServer).(string)
	return bwscli.NewSecretsManagerClient(serverURL), nil
}

func buildWebapiOptions(version string) []webapi.Options {
	webapiOpts := []webapi.Options{}
	if version == versionTestDisabledRetries {
		// During development, we don't want to wait on any sporadic errors.
		webapiOpts = append(webapiOpts, webapi.DisableRetries())
	}
	return webapiOpts
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

func ensureLoggedInEmbeddedSecretsManager(ctx context.Context, d *schema.ResourceData, bwClient embedded.SecretsManager) error {
	accessToken, hasAccessToken := d.GetOk(schema_definition.AttributeAccessToken)
	if !hasAccessToken {
		return fmt.Errorf("access token is required")
	}

	loginMethod := loginMethod(d)
	switch loginMethod {
	case LoginMethodAccessToken:
		return bwClient.LoginWithAccessToken(ctx, accessToken.(string))
	}

	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: 'BUG')")
}

func ensureLoggedInEmbeddedPasswordManager(ctx context.Context, d *schema.ResourceData, bwClient bitwarden.PasswordManager) error {
	masterPassword, hasMasterPassword := d.GetOk(schema_definition.AttributeMasterPassword)
	if !hasMasterPassword {
		return fmt.Errorf("master password is required")
	}

	loginMethod := loginMethod(d)
	switch loginMethod {
	case LoginMethodPersonalAPIKey:
		clientID := d.Get(schema_definition.AttributeClientID)
		clientSecret := d.Get(schema_definition.AttributeClientSecret)
		return bwClient.LoginWithAPIKey(ctx, masterPassword.(string), clientID.(string), clientSecret.(string))
	case LoginMethodPassword:
		email := d.Get(schema_definition.AttributeProviderEmail)
		return bwClient.LoginWithPassword(ctx, email.(string), masterPassword.(string))
	}

	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: 'BUG')")
}

func hasExperimentalFeature(d *schema.ResourceData, feature string) bool {
	experimentalFeatures, hasExperimentalFeatures := d.GetOk(schema_definition.AttributeExperimental)
	if !hasExperimentalFeatures {
		return false
	}
	if experimentalFeatures.(*schema.Set).Len() == 0 {
		return false
	}

	return experimentalFeatures.(*schema.Set).List()[0].(map[string]interface{})[feature].(bool)
}
