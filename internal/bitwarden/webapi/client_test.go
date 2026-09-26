//go:build offline

package webapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonObject struct {
	ID string `json:"id"`
}

func TestDoRequest_EmptyJSONBodyReturnsError(t *testing.T) {
	res, err := doTestRequest[jsonObject](t, nil)

	assert.Nil(t, res)
	require.ErrorContains(t, err, "empty response body")
}

func TestDoRequest_EmptyByteBodySucceeds(t *testing.T) {
	res, err := doTestRequest[[]byte](t, nil)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Empty(t, *res)
}

func TestDoRequest_MalformedJSONReturnsError(t *testing.T) {
	const body = `not-json`
	res, err := doTestRequest[jsonObject](t, []byte(body))

	assert.Nil(t, res)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), body)
}

func doTestRequest[T any](t *testing.T, body []byte) (*T, error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	return doRequest[T](t.Context(), server.Client(), req)
}
