//go:build offline

package webapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
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

func TestDoRequest_HTTPErrorUsesRecognizedMessage(t *testing.T) {
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(`{"object":"error","message":"cipher not found"}`))

	assert.Nil(t, res)
	require.EqualError(t, err, `the server returned an error: "cipher not found" (400)`)
}

func TestDoRequest_TraceLogsOmitBodiesAndAuthorization(t *testing.T) {
	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &logs)

	const token = "Bearer secret-token"
	const reqSecret = `{"password":"super-secret-password"}`
	const respSecret = `{"id":"ok","access_token":"leaked-access-token"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "session=leaked-session-cookie")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(respSecret))
	}))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/identity/connect/token?code=oauth-secret", strings.NewReader(reqSecret))
	require.NoError(t, err)
	req.Header.Set("Authorization", token)

	res, err := doRequest[jsonObject](ctx, server.Client(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "ok", res.ID)

	logOut := logs.String()
	assert.Contains(t, logOut, "/identity/connect/token")
	assert.NotContains(t, logOut, "secret-token")
	assert.NotContains(t, logOut, "super-secret-password")
	assert.NotContains(t, logOut, "leaked-access-token")
	assert.NotContains(t, logOut, "leaked-session-cookie")
	assert.NotContains(t, logOut, "oauth-secret")
	assert.NotContains(t, logOut, "Authorization")
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

func doTestErrorRequest[T any](t *testing.T, status int, body []byte) (*T, error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	return doRequest[T](t.Context(), server.Client(), req)
}
