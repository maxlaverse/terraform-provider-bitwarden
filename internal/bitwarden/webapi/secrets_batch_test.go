//go:build offline

package webapi

import (
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecretsByIDsMalformedResponseDoesNotPrint(t *testing.T) {
	transport := httpmock.NewMockTransport()
	transport.RegisterResponder("POST", "http://127.0.0.1/api/secrets/get-by-ids",
		httpmock.NewStringResponder(http.StatusOK, `{"data": "private-response-value"`))
	client := &client{
		serverURL:          "http://127.0.0.1",
		sessionAccessToken: "test-token",
		httpClient:         &http.Client{Transport: transport},
	}

	output, err := os.CreateTemp(t.TempDir(), "stdout")
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = output
	t.Cleanup(func() {
		os.Stdout = stdout
		_ = output.Close()
	})

	_, err = client.GetSecretsByIDs(t.Context(), []string{"secret-id"})
	os.Stdout = stdout

	require.Error(t, err)
	data, readErr := os.ReadFile(output.Name())
	require.NoError(t, readErr)
	assert.Empty(t, string(data), "API response bodies must not be printed to provider stdout")
	assert.NotContains(t, err.Error(), "private-response-value")
}
