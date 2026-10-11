//go:build offline

package webapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPErrorFromResponse(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.com/identity/connect/token", nil)
	require.NoError(t, err)

	tests := []struct {
		name string
		body string
		want string
		omit string
	}{
		{
			name: "api error envelope",
			body: `{"object":"error","message":"cipher not found"}`,
			want: `the server returned an error: "cipher not found" (400)`,
		},
		{
			name: "vaultwarden login envelope with validationErrors object",
			body: `{"message":"Username or password is incorrect. Try again","error":"","error_description":"","validationErrors":{"":["Username or password is incorrect. Try again"]},"errorModel":{"message":"Username or password is incorrect. Try again","object":"error"},"object":"error"}`,
			want: `the server returned an error: "Username or password is incorrect. Try again" (400)`,
		},
		{
			name: "oauth description",
			body: `{"error":"invalid_grant","error_description":"Username or password is incorrect"}`,
			want: "Username or password is incorrect",
			omit: `"error"`,
		},
		{
			name: "oauth ErrorModel",
			body: `{"error":"invalid_grant","ErrorModel":{"Message":"Username or password is incorrect","Object":"error"}}`,
			want: "Username or password is incorrect",
		},
		{
			name: "oauth code only",
			body: `{"error":"invalid_scope"}`,
			want: "invalid_scope",
		},
		{
			name: "unrecognized json omits body",
			body: `{"detail":"THIS_IS_A_SECRET"}`,
			want: "bad response status code for 'POST http://example.com/identity/connect/token': 400!=200",
			omit: "THIS_IS_A_SECRET",
		},
		{
			name: "non-json omits body",
			body: "attachment not found",
			want: "bad response status code for 'POST http://example.com/identity/connect/token': 400!=200",
			omit: "attachment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: http.StatusBadRequest}
			got := httpErrorFromResponse(req, resp, []byte(tt.body))
			require.EqualError(t, got, tt.want)
			if tt.omit != "" {
				assert.NotContains(t, got.Error(), tt.omit)
			}
		})
	}
}
