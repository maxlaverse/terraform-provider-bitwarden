//go:build offline

package provider

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/bwcli"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type vaultPathAppDataCase struct {
	name           string
	vaultPath      *string
	wantAppDataDir bool
	expectedDir    string
}

func vaultPathAppDataCases(t *testing.T) []vaultPathAppDataCase {
	t.Helper()

	custom := "custom_app_dir"
	customAbs, err := filepath.Abs(custom)
	require.NoError(t, err)
	defaultAbs, err := filepath.Abs(".bitwarden/")
	require.NoError(t, err)
	empty := ""

	return []vaultPathAppDataCase{
		{
			name:           "omitted uses .bitwarden",
			vaultPath:      nil,
			wantAppDataDir: true,
			expectedDir:    defaultAbs,
		},
		{
			name:           "empty omits BITWARDENCLI_APPDATA_DIR",
			vaultPath:      &empty,
			wantAppDataDir: false,
		},
		{
			name:           "set uses provided path",
			vaultPath:      &custom,
			wantAppDataDir: true,
			expectedDir:    customAbs,
		},
	}
}

func TestCLIAppDataDirFromVaultPath_FrameworkConfigure(t *testing.T) {
	t.Setenv("BITWARDENCLI_APPDATA_DIR", "")

	for _, tc := range vaultPathAppDataCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			pm := configureFrameworkWithVaultPath(t, tc.vaultPath)
			assertCLIAppDataDir(t, pm, tc.wantAppDataDir, tc.expectedDir)
		})
	}
}

func TestCLIAppDataDirFromVaultPath_SDKConfigure(t *testing.T) {
	t.Setenv("BITWARDENCLI_APPDATA_DIR", "")

	for _, tc := range vaultPathAppDataCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			pm := configureSDKWithVaultPath(t, tc.vaultPath)
			assertCLIAppDataDir(t, pm, tc.wantAppDataDir, tc.expectedDir)
		})
	}
}

func TestApplyProviderConfigEnvDefaults_VaultPath(t *testing.T) {
	t.Run("omitted uses environment", func(t *testing.T) {
		t.Setenv("BITWARDENCLI_APPDATA_DIR", "/custom/cli-dir")
		cfg := applyProviderConfigEnvDefaults(providerConfig{})
		assert.Equal(t, explicitVaultPath("/custom/cli-dir"), cfg.VaultPath)
	})

	t.Run("explicit empty skips environment", func(t *testing.T) {
		t.Setenv("BITWARDENCLI_APPDATA_DIR", "/custom/cli-dir")
		cfg := applyProviderConfigEnvDefaults(providerConfig{VaultPath: explicitVaultPath("")})
		assert.Equal(t, explicitVaultPath(""), cfg.VaultPath)
	})
}

func configureFrameworkWithVaultPath(t *testing.T, vaultPath *string) bitwarden.PasswordManager {
	t.Helper()
	ctx := t.Context()
	p := New(versionTestSkippedLogin)()

	var schemaResp fwprovider.SchemaResponse
	p.Schema(ctx, fwprovider.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), schemaResp.Diagnostics)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, nil)
	}
	attrs[schema_definition.AttributeSessionKey] = tftypes.NewValue(tftypes.String, t.Name())
	if vaultPath != nil {
		attrs[schema_definition.AttributeVaultPath] = tftypes.NewValue(tftypes.String, *vaultPath)
	}

	var resp fwprovider.ConfigureResponse
	p.Configure(ctx, fwprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(objType, attrs),
			Schema: schemaResp.Schema,
		},
	}, &resp)
	require.False(t, resp.Diagnostics.HasError(), fmt.Sprintf("%v", resp.Diagnostics))

	clients, ok := resp.ResourceData.(*ProviderClients)
	require.True(t, ok)
	pm, err := clients.RequirePasswordManager()
	require.NoError(t, err)
	return pm
}

func configureSDKWithVaultPath(t *testing.T, vaultPath *string) bitwarden.PasswordManager {
	t.Helper()
	sdk := NewSDK(versionTestSkippedLogin)()

	raw := map[string]interface{}{
		schema_definition.AttributeSessionKey: t.Name(),
	}
	if vaultPath != nil {
		raw[schema_definition.AttributeVaultPath] = *vaultPath
	}

	typ := schema.InternalMap(sdk.Schema).CoreConfigSchema().ImpliedType()
	vals := make(map[string]cty.Value, len(typ.AttributeTypes()))
	for name, at := range typ.AttributeTypes() {
		vals[name] = cty.NullVal(at)
	}
	vals[schema_definition.AttributeSessionKey] = cty.StringVal(t.Name())
	if vaultPath != nil {
		vals[schema_definition.AttributeVaultPath] = cty.StringVal(*vaultPath)
	}

	cfg := terraform.NewResourceConfigRaw(raw)
	cfg.CtyValue = cty.ObjectVal(vals)

	diags := sdk.Configure(t.Context(), cfg)
	require.False(t, diags.HasError(), diags)

	clients, ok := sdk.Meta().(*ProviderClients)
	require.True(t, ok)
	pm, err := clients.RequirePasswordManager()
	require.NoError(t, err)
	return pm
}

func assertCLIAppDataDir(t *testing.T, pm bitwarden.PasswordManager, wantSet bool, wantDir string) {
	t.Helper()
	env, ok := bwcli.CLIEnv(pm)
	require.True(t, ok, "expected a CLI password manager client")

	dir, set := cliAppDataDir(env)
	if !wantSet {
		assert.False(t, set, "BITWARDENCLI_APPDATA_DIR should not be set, got %q", dir)
		return
	}
	require.True(t, set, "BITWARDENCLI_APPDATA_DIR should be set")
	assert.Equal(t, wantDir, dir)
}

func cliAppDataDir(env []string) (string, bool) {
	const prefix = "BITWARDENCLI_APPDATA_DIR="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix), true
		}
	}
	return "", false
}
