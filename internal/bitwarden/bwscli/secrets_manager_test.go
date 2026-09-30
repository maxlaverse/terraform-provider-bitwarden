//go:build offline

package bwscli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	test_command "github.com/maxlaverse/terraform-provider-bitwarden/internal/command/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecretsByIDs(t *testing.T) {
	var ids []string
	var wanted []models.Secret
	for i := range 77 {
		id := fmt.Sprintf("secret-%02d", i)
		ids = append(ids, id)
		wanted = append(wanted, models.Secret{ID: id, Value: fmt.Sprintf("private-value-%d\"\n雪", i)})
	}
	wanted[0].Value = ""
	ids = append(ids, strings.ToUpper(ids[1]))
	listed := append(slices.Clone(wanted), models.Secret{ID: "unrelated-id", Value: "unrelated-private"})
	slices.Reverse(listed)
	data, err := json.Marshal(listed)
	require.NoError(t, err)

	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"secret list --output json": string(data),
	})
	defer removeMocks(t)

	client := NewSecretsManagerClient("http://127.0.0.1")
	require.NoError(t, client.LoginWithAccessToken(t.Context(), "test-token"))

	got, err := client.GetSecretsByIDs(t.Context(), ids)

	require.NoError(t, err)
	assert.Equal(t, []string{"secret list --output json"}, commandsExecuted())
	assert.ElementsMatch(t, wanted, got, "only the requested IDs should be returned, once each")
}

func TestGetSecretsByIDsEmpty(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, nil)
	defer removeMocks(t)

	client := NewSecretsManagerClient("http://127.0.0.1")
	require.NoError(t, client.LoginWithAccessToken(t.Context(), "test-token"))

	got, err := client.GetSecretsByIDs(t.Context(), nil)

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Empty(t, commandsExecuted(), "an empty batch must not run the CLI")
}

func TestGetSecretsByIDsErrors(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected error
	}{
		{
			name:     "missing ID",
			output:   `[{"id":"a","value":"private-value"}]`,
			expected: models.ErrObjectNotFound,
		},
		{
			name:     "empty list",
			output:   `[]`,
			expected: models.ErrObjectNotFound,
		},
		{
			name:     "null list",
			output:   `null`,
			expected: models.ErrObjectNotFound,
		},
		{
			name:     "duplicate ID",
			output:   `[{"id":"a"},{"id":"A"},{"id":"b"}]`,
			expected: models.ErrTooManyObjectsFound,
		},
		{
			name:   "invalid JSON",
			output: `[{"id":"a","value":"private-value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
				"secret list --output json": tt.output,
			})
			defer removeMocks(t)

			client := NewSecretsManagerClient("http://127.0.0.1")
			require.NoError(t, client.LoginWithAccessToken(t.Context(), "test-token"))

			got, err := client.GetSecretsByIDs(t.Context(), []string{"a", "b"})

			require.Error(t, err)
			if tt.expected != nil {
				assert.ErrorIs(t, err, tt.expected)
			}
			assert.Nil(t, got, "errors must not return a partial batch")
			assert.NotContains(t, err.Error(), "private-value")
			assert.Equal(t, []string{"secret list --output json"}, commandsExecuted())
		})
	}
}

func TestGetSecretsByIDsWithoutAccessToken(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, nil)
	defer removeMocks(t)

	client := NewSecretsManagerClient("http://127.0.0.1")

	got, err := client.GetSecretsByIDs(t.Context(), []string{"a"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Empty(t, commandsExecuted(), "an unauthenticated client must not run the CLI")
}
