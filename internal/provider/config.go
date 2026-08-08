package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

// providerConfig is a client-agnostic representation of the provider block. It
// is populated from either the Plugin Framework model or SDKv2 ResourceData
// (with environment-variable fallbacks applied) and consumed by
// configureClients. Empty strings mean "not set" for all string fields.
type providerConfig struct {
	Server                                        string
	Email                                         string
	MasterPassword                                string
	SessionKey                                    string
	ClientID                                      string
	ClientSecret                                  string
	AccessToken                                   string
	VaultPath                                     string
	ExtraCACertsPath                              string
	ClientImplementation                          string
	ExperimentalEmbeddedClient                    bool
	ExperimentalDisableSyncAfterWriteVerification bool
}

func (c providerConfig) has(value string) bool { return len(value) > 0 }

// cacheKey identifies equivalent provider configurations so the Framework and
// SDKv2 halves of a single mux ConfigureProvider RPC can hand off one client.
func (c providerConfig) cacheKey(version string) string {
	return strings.Join([]string{
		version,
		c.Server,
		c.Email,
		c.MasterPassword,
		c.SessionKey,
		c.ClientID,
		c.ClientSecret,
		c.AccessToken,
		c.VaultPath,
		c.ExtraCACertsPath,
		c.ClientImplementation,
		fmt.Sprintf("%t", c.ExperimentalEmbeddedClient),
		fmt.Sprintf("%t", c.ExperimentalDisableSyncAfterWriteVerification),
	}, "\x00")
}

var (
	muxClientsMu    sync.Mutex
	muxClientsOffer = map[string]*ProviderClients{}
)

// configureClientsOffer builds clients for the Framework mux half and parks
// them for the following SDKv2 Configure call. A permanent process-wide cache
// is intentionally avoided: embedded clients keep an in-memory vault that must
// be rebuilt on later ConfigureProvider RPCs (e.g. after an external delete).
func configureClientsOffer(ctx context.Context, version string, cfg providerConfig) (*ProviderClients, error) {
	clients, err := configureClients(ctx, version, cfg)
	if err != nil {
		return nil, err
	}

	key := cfg.cacheKey(version)
	muxClientsMu.Lock()
	muxClientsOffer[key] = clients
	muxClientsMu.Unlock()
	return clients, nil
}

// configureClientsTakeOrCreate returns clients parked by configureClientsOffer
// for this config, or builds a fresh pair when nothing was offered (Framework
// Configure skipped, or a different config key).
func configureClientsTakeOrCreate(ctx context.Context, version string, cfg providerConfig) (*ProviderClients, error) {
	key := cfg.cacheKey(version)

	muxClientsMu.Lock()
	if offered, ok := muxClientsOffer[key]; ok {
		delete(muxClientsOffer, key)
		muxClientsMu.Unlock()
		return offered, nil
	}
	muxClientsMu.Unlock()

	return configureClients(ctx, version, cfg)
}

// providerConfigureSDK adapts SDKv2 ResourceData into providerConfig and
// reuses configureClients so both muxed providers share one login path.
func providerConfigureSDK(version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		cfg := applyProviderConfigEnvDefaults(providerConfigFromResourceData(d))
		if err := validateProviderConfig(cfg); err != nil {
			return nil, diag.Errorf("%s", err.Error())
		}
		// Mux calls Framework Configure first; take its clients when present so
		// we do not log in twice for the same ConfigureProvider RPC.
		clients, err := configureClientsTakeOrCreate(ctx, version, cfg)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return clients, nil
	}
}

func providerConfigFromResourceData(d *schema.ResourceData) providerConfig {
	cfg := providerConfig{
		Server:               stringFromResourceData(d, schema_definition.AttributeServer),
		Email:                stringFromResourceData(d, schema_definition.AttributeProviderEmail),
		MasterPassword:       stringFromResourceData(d, schema_definition.AttributeMasterPassword),
		SessionKey:           stringFromResourceData(d, schema_definition.AttributeSessionKey),
		ClientID:             stringFromResourceData(d, schema_definition.AttributeClientID),
		ClientSecret:         stringFromResourceData(d, schema_definition.AttributeClientSecret),
		AccessToken:          stringFromResourceData(d, schema_definition.AttributeBwsAccessToken),
		VaultPath:            stringFromResourceData(d, schema_definition.AttributeVaultPath),
		ExtraCACertsPath:     stringFromResourceData(d, schema_definition.AttributeExtraCACertsPath),
		ClientImplementation: stringFromResourceData(d, schema_definition.AttributeClientImplementation),
	}

	if experimental, ok := d.GetOk(schema_definition.AttributeExperimental); ok {
		set := experimental.(*schema.Set)
		if set.Len() > 0 {
			m := set.List()[0].(map[string]interface{})
			if v, ok := m[schema_definition.AttributeExperimentalEmbeddedClient].(bool); ok {
				cfg.ExperimentalEmbeddedClient = v
			}
			if v, ok := m[schema_definition.AttributeExperimentalDisableSyncAfterWriteVerification].(bool); ok {
				cfg.ExperimentalDisableSyncAfterWriteVerification = v
			}
		}
	}

	return cfg
}

func stringFromResourceData(d *schema.ResourceData, key string) string {
	if s, ok := d.Get(key).(string); ok {
		return s
	}
	return ""
}

// applyProviderConfigEnvDefaults fills empty config fields from the environment
// (and hard-coded defaults), replacing SDKv2 schema DefaultFunc behaviour.
//
// Note: an explicitly empty vault_path currently cannot be distinguished from
// "unset", so "" still falls through to BITWARDENCLI_APPDATA_DIR / .bitwarden/.
func applyProviderConfigEnvDefaults(cfg providerConfig) providerConfig {
	cfg.Server = firstNonEmpty(cfg.Server, envFirst("BW_URL", "BWS_SERVER_URL"), bitwarden.DefaultBitwardenServerURL)
	cfg.Email = firstNonEmpty(cfg.Email, envFirst("BW_EMAIL"))
	cfg.MasterPassword = firstNonEmpty(cfg.MasterPassword, envFirst("BW_PASSWORD"))
	cfg.SessionKey = firstNonEmpty(cfg.SessionKey, envFirst("BW_SESSION"))
	cfg.ClientID = firstNonEmpty(cfg.ClientID, envFirst("BW_CLIENTID"))
	cfg.ClientSecret = firstNonEmpty(cfg.ClientSecret, envFirst("BW_CLIENTSECRET"))
	cfg.AccessToken = firstNonEmpty(cfg.AccessToken, envFirst("BWS_ACCESS_TOKEN"))
	cfg.VaultPath = firstNonEmpty(cfg.VaultPath, envFirst("BITWARDENCLI_APPDATA_DIR"), ".bitwarden/")
	cfg.ExtraCACertsPath = firstNonEmpty(cfg.ExtraCACertsPath, envFirst("NODE_EXTRA_CA_CERTS"))
	return cfg
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if len(v) > 0 {
			return v
		}
	}
	return ""
}

func envFirst(keys ...string) string {
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
	}
	return ""
}

// validateProviderConfig re-implements the cross-field credential rules that
// used to live in the SDKv2 schema (ConflictsWith/RequiredWith/AtLeastOneOf).
// Both muxed providers call it from Configure after env defaults are applied.
func validateProviderConfig(cfg providerConfig) error {
	hasMasterPassword := cfg.has(cfg.MasterPassword)
	hasSessionKey := cfg.has(cfg.SessionKey)
	hasAccessToken := cfg.has(cfg.AccessToken)
	hasClientID := cfg.has(cfg.ClientID)
	hasClientSecret := cfg.has(cfg.ClientSecret)
	hasEmail := cfg.has(cfg.Email)

	if !hasMasterPassword && !hasSessionKey && !hasAccessToken {
		return fmt.Errorf("one of `access_token`, `master_password` or `session_key` must be specified")
	}

	if hasMasterPassword && (hasSessionKey || hasAccessToken) {
		return fmt.Errorf("`master_password` conflicts with `session_key` and `access_token`")
	}

	if hasClientID != hasClientSecret {
		return fmt.Errorf("`client_id` and `client_secret` must be specified together")
	}

	if (hasClientID || hasClientSecret) && !hasMasterPassword {
		return fmt.Errorf("`client_id` and `client_secret` require `master_password` to also be specified")
	}

	if !hasEmail && !hasAccessToken && !hasClientID && !hasSessionKey {
		return fmt.Errorf("one of `access_token`, `client_id`, `email` or `session_key` must be specified")
	}

	return nil
}

// configureClients builds the Bitwarden clients described by the provider
// configuration. It mirrors the behaviour of the historical SDKv2
// providerConfigure implementation.
func configureClients(ctx context.Context, version string, cfg providerConfig) (*ProviderClients, error) {
	shouldLogin := !strings.Contains(version, versionTestSkippedLogin)

	hasAccessToken := cfg.has(cfg.AccessToken)
	useEmbeddedClient := getClientImplementation(cfg) == schema_definition.ClientImplementationEmbedded

	if useEmbeddedClient && cfg.has(cfg.SessionKey) {
		return nil, fmt.Errorf("session key is not supported with the embedded client")
	}

	if useEmbeddedClient && !hasAccessToken {
		bwClient, err := newEmbeddedPasswordManagerClient(ctx, cfg, version)
		if err != nil {
			return nil, err
		}

		if shouldLogin {
			if err = ensureLoggedInEmbeddedPasswordManager(ctx, cfg, bwClient); err != nil {
				return nil, err
			}
		}
		return &ProviderClients{PasswordManager: bwClient}, nil
	} else if useEmbeddedClient && hasAccessToken {
		bwsClient, err := newEmbeddedSecretsManagerClient(ctx, cfg, version)
		if err != nil {
			return nil, err
		}

		if shouldLogin {
			if err = ensureLoggedInEmbeddedSecretsManager(ctx, cfg, bwsClient); err != nil {
				return nil, err
			}
		}
		return &ProviderClients{SecretsManager: bwsClient}, nil
	} else if !useEmbeddedClient && hasAccessToken {
		bwsClient, err := newCLISecretsManagerClient(ctx, cfg, version)
		if err != nil {
			return nil, err
		}

		// We login anyway, since it's just about storing the access token
		// when using the CLI.
		if err = ensureLoggedInEmbeddedSecretsManager(ctx, cfg, bwsClient); err != nil {
			return nil, err
		}

		return &ProviderClients{SecretsManager: bwsClient}, nil
	}

	bwClient, err := newCLIPasswordManagerClient(cfg, version)
	if err != nil {
		return nil, err
	}

	if cfg.has(cfg.SessionKey) {
		bwClient.SetSessionKey(cfg.SessionKey)
	}

	if shouldLogin {
		if err = ensureLoggedInCLIPasswordManager(ctx, cfg, bwClient); err != nil {
			return nil, err
		}
	}

	return &ProviderClients{PasswordManager: bwClient}, nil
}

func getClientImplementation(cfg providerConfig) string {
	// Check for backward compatibility with experimental.embedded_client.
	// When both are set, experimental.embedded_client wins. We do not reject the
	// combination: client_implementation defaults to "cli" here when unset, so a
	// schema ConflictsWith cannot tell "explicit cli" from "defaulted".
	if cfg.ExperimentalEmbeddedClient {
		tflog.Warn(context.Background(), "The experimental.embedded_client attribute is deprecated. Please use client_implementation = \"embedded\" instead.")
		return schema_definition.ClientImplementationEmbedded
	}

	// Get client_implementation, defaulting to "cli" if not set.
	if !cfg.has(cfg.ClientImplementation) {
		return schema_definition.ClientImplementationCLI
	}
	return cfg.ClientImplementation
}

func ensureLoggedInCLIPasswordManager(ctx context.Context, cfg providerConfig, bwClient bwcli.PasswordManagerClient) error {
	status, err := bwClient.Status(ctx)
	if err != nil {
		return err
	}

	if err = logoutIfIdentityChanged(ctx, cfg, bwClient, status); err != nil {
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
	if cfg.has(cfg.MasterPassword) && status.Status == bwcli.StatusLocked {
		if err = bwClient.Unlock(ctx, cfg.MasterPassword); err != nil {
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
	switch loginMethod(cfg) {
	case LoginMethodPersonalAPIKey:
		return bwClient.LoginWithAPIKey(ctx, cfg.MasterPassword, cfg.ClientID, cfg.ClientSecret)
	case LoginMethodPassword:
		return bwClient.LoginWithPassword(ctx, cfg.Email, cfg.MasterPassword)
	}

	// Scenario 4: We need to login but don't have the information to do so.
	//             This is a situation we can't get out from.
	//             => failure
	if cfg.has(cfg.SessionKey) {
		return fmt.Errorf("unable to unlock Vault with provided session key (status: %s)", status.Status)
	}

	// We should have caught already scenarios up to this point. If we haven't, it means this method's
	// implementation is wrong or the provider parameters are.
	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: '%s')", status.Status)
}

func loginMethod(cfg providerConfig) LoginMethod {
	if cfg.has(cfg.AccessToken) {
		return LoginMethodAccessToken
	} else if cfg.has(cfg.ClientID) && cfg.has(cfg.ClientSecret) {
		return LoginMethodPersonalAPIKey
	} else if cfg.has(cfg.MasterPassword) {
		return LoginMethodPassword
	}

	return LoginMethodNone
}

func logoutIfIdentityChanged(ctx context.Context, cfg providerConfig, bwClient bwcli.PasswordManagerClient, status *bwcli.Status) error {
	serverURL := cfg.Server
	emailProvided := cfg.has(cfg.Email)
	vaultBelongsToEmailAndServer := (!emailProvided || status.VaultOfUser(cfg.Email)) && status.VaultFromServer(serverURL)

	if (status.Status == bwcli.StatusLocked || status.Status == bwcli.StatusUnlocked) && !vaultBelongsToEmailAndServer {
		status.Status = bwcli.StatusUnauthenticated

		tflog.Warn(ctx, "Logging out as the local Vault belongs to a different identity", map[string]interface{}{"vault_email": status.UserEmail, "vault_server": status.ServerURL, "provider_server": serverURL})
		if err := bwClient.Logout(ctx); err != nil {
			return err
		}
	}

	if !status.VaultFromServer(serverURL) {
		if err := bwClient.SetServer(ctx, serverURL); err != nil {
			return err
		}
	}
	return nil
}

func newCLIPasswordManagerClient(cfg providerConfig, version string) (bwcli.PasswordManagerClient, error) {
	opts := []bwcli.Options{}
	if cfg.has(cfg.VaultPath) {
		abs, err := filepath.Abs(cfg.VaultPath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, bwcli.WithAppDataDir(abs))
	}

	if cfg.has(cfg.ExtraCACertsPath) {
		opts = append(opts, bwcli.WithExtraCACertsPath(cfg.ExtraCACertsPath))
	}

	if version == versionTestDisabledRetries {
		// During development, we disable retry backoffs to make some operations faster.
		opts = append(opts, bwcli.DisableRetryBackoff())
	}

	return bwcli.NewPasswordManagerClient(opts...), nil
}

func newEmbeddedPasswordManagerClient(ctx context.Context, cfg providerConfig, version string) (bitwarden.PasswordManager, error) {
	deviceId, err := getOrGenerateDeviceIdentifier(ctx)
	if err != nil {
		return nil, err
	}

	opts := []embedded.PasswordManagerOptions{
		embedded.WithPasswordManagerHttpOptions(buildWebapiOptions(version)...),
	}

	if cfg.ExperimentalDisableSyncAfterWriteVerification {
		opts = append(opts, embedded.DisableFailOnSyncAfterWriteVerification())
	}

	return embedded.NewPasswordManagerClient(cfg.Server, deviceId, version, opts...), nil
}

func newEmbeddedSecretsManagerClient(ctx context.Context, cfg providerConfig, version string) (bitwarden.SecretsManager, error) {
	deviceId, err := getOrGenerateDeviceIdentifier(ctx)
	if err != nil {
		return nil, err
	}

	return embedded.NewSecretsManagerClient(cfg.Server, deviceId, version, embedded.WithSecretsManagerHttpOptions(buildWebapiOptions(version)...)), nil
}

func newCLISecretsManagerClient(_ context.Context, cfg providerConfig, _ string) (bitwarden.SecretsManager, error) {
	return bwscli.NewSecretsManagerClient(cfg.Server), nil
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

func ensureLoggedInEmbeddedSecretsManager(ctx context.Context, cfg providerConfig, bwClient embedded.SecretsManager) error {
	if !cfg.has(cfg.AccessToken) {
		return fmt.Errorf("access token is required")
	}

	switch loginMethod(cfg) {
	case LoginMethodAccessToken:
		return bwClient.LoginWithAccessToken(ctx, cfg.AccessToken)
	}

	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: 'BUG')")
}

func ensureLoggedInEmbeddedPasswordManager(ctx context.Context, cfg providerConfig, bwClient bitwarden.PasswordManager) error {
	if !cfg.has(cfg.MasterPassword) {
		return fmt.Errorf("master password is required")
	}

	switch loginMethod(cfg) {
	case LoginMethodPersonalAPIKey:
		return bwClient.LoginWithAPIKey(ctx, cfg.MasterPassword, cfg.ClientID, cfg.ClientSecret)
	case LoginMethodPassword:
		return bwClient.LoginWithPassword(ctx, cfg.Email, cfg.MasterPassword)
	}

	return fmt.Errorf("INTERNAL BUG: not enough parameters provided to login (status: 'BUG')")
}
