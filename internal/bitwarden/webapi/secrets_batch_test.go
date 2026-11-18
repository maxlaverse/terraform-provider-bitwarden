//go:build offline

package webapi

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecretsByIDsRetriesInternalServerError(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	attempts := 0
	retry := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	retry.Transport = &roundTripFunc{fn: func(req *http.Request) (*http.Response, error) {
		attempts++
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/api/secrets/get-by-ids", req.URL.Path)
		require.NotNil(t, req.Body)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"ids":["first","second"]}`, string(body))

		if attempts == 1 {
			return &http.Response{StatusCode: http.StatusInternalServerError}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"id":"first","value":"encrypted-value"},{"id":"second"}]}`)),
		}, nil
	}}

	client := &client{
		serverURL:          "http://127.0.0.1",
		sessionAccessToken: "test-token",
		httpClient:         &http.Client{Transport: retry},
	}

	secrets, err := client.GetSecretsByIDs(t.Context(), []string{"first", "second"})
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "first", secrets[0].ID)
	assert.Equal(t, "encrypted-value", secrets[0].Value)
	assert.Equal(t, 2, attempts)
}
