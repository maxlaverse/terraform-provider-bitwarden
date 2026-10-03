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

func TestDoRequest_APIErrorResponseUsesMessage(t *testing.T) {
	const body = `{"object":"error","message":"cipher not found"}`
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(body))

	assert.Nil(t, res)
	require.EqualError(t, err, `the server returned an error: "cipher not found" (400)`)
}

func TestDoRequest_APIErrorResponseWithValidationErrorsObject(t *testing.T) {
	const body = `{"message":"Username or password is incorrect. Try again","error":"","error_description":"","validationErrors":{"":["Username or password is incorrect. Try again"]},"errorModel":{"message":"Username or password is incorrect. Try again","object":"error"},"object":"error"}`
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(body))

	assert.Nil(t, res)
	require.EqualError(t, err, `the server returned an error: "Username or password is incorrect. Try again" (400)`)
}

func TestDoRequest_OAuthTokenErrorUsesDescription(t *testing.T) {
	const body = `{"error":"invalid_grant","error_description":"Username or password is incorrect"}`
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(body))

	assert.Nil(t, res)
	require.EqualError(t, err, "Username or password is incorrect")
	assert.NotContains(t, err.Error(), `"error"`)
}

func TestDoRequest_OAuthTokenErrorFallsBackToErrorModel(t *testing.T) {
	const body = `{"error":"invalid_grant","ErrorModel":{"Message":"Username or password is incorrect","Object":"error"}}`
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(body))

	assert.Nil(t, res)
	require.EqualError(t, err, "Username or password is incorrect")
}

func TestDoRequest_OAuthTokenErrorFallsBackToCode(t *testing.T) {
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(`{"error":"invalid_scope"}`))

	assert.Nil(t, res)
	require.EqualError(t, err, "invalid_scope")
}

func TestDoRequest_NonOAuthErrorOmitsBody(t *testing.T) {
	const body = `{"detail":"THIS_IS_A_SECRET"}`
	res, err := doTestErrorRequest[jsonObject](t, http.StatusBadRequest, []byte(body))

	assert.Nil(t, res)
	require.ErrorContains(t, err, "400!=200")
	assert.NotContains(t, err.Error(), "THIS_IS_A_SECRET")
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	return doRequest[T](t.Context(), server.Client(), req)
}
