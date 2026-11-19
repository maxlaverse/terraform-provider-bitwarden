//go:build offline

package webapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigResponseServerType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
		want ServerType
	}{
		{
			name: "official bitwarden omits server",
			json: `{"version":"2025.6.2","gitHash":"b511b055","server":null,"object":"config"}`,
			want: ServerTypeBitwarden,
		},
		{
			name: "official bitwarden without server field",
			json: `{"version":"2025.6.2","featureStates":{"cipher-key-encryption":false}}`,
			want: ServerTypeBitwarden,
		},
		{
			name: "vaultwarden identifies itself",
			json: `{
				"version":"2026.6.0",
				"gitHash":"cdfdc6ff",
				"server":{"name":"Vaultwarden","url":"https://github.com/dani-garcia/vaultwarden"},
				"object":"config"
			}`,
			want: ServerTypeVaultwarden,
		},
		{
			name: "other unofficial servers default to bitwarden",
			json: `{"server":{"name":"Third Party Bitwarden Server","url":"https://example.com"}}`,
			want: ServerTypeBitwarden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cfg ConfigResponse
			require.NoError(t, json.Unmarshal([]byte(tt.json), &cfg))
			assert.Equal(t, tt.want, cfg.ServerType())
		})
	}

	assert.Equal(t, ServerTypeBitwarden, (*ConfigResponse)(nil).ServerType())
}
