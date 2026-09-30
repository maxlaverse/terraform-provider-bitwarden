//go:build offline

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSecretsEmbeddedBatch(t *testing.T) {
	const secretCount = 3
	backend := NewTestSecretsManager()
	router := mux.NewRouter()
	router.HandleFunc("/identity/connect/token", backend.handlerLogin).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/projects", backend.handlerCreateProject).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/secrets", backend.handlerCreateGetSecret).Methods("POST")

	batchCalls := 0
	outcome := "success"
	router.HandleFunc("/api/secrets/get-by-ids", func(w http.ResponseWriter, req *http.Request) {
		batchCalls++
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		var body struct {
			IDs []string `json:"ids"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Len(t, body.IDs, secretCount, "different casing must not create duplicate API IDs")
		response := webapi.SecretsList{Data: make([]webapi.Secret, 0, len(body.IDs))}
		for _, id := range body.IDs {
			secret, ok := backend.secretsStore[id]
			require.True(t, ok)
			// The bulk endpoint can omit projects. Do not rely on Projects[0].
			secret.Projects = nil
			response.Data = append(response.Data, secret)
		}
		slices.Reverse(response.Data)
		w.Header().Set("Content-Type", "application/json")
		switch outcome {
		case "missing":
			response.Data = response.Data[:len(response.Data)-1]
		case "duplicate":
			response.Data[0] = response.Data[1]
		case "unexpected":
			response.Data[0].ID = "unexpected-id"
		case "corrupt ciphertext":
			response.Data[0].Value = "invalid-private-value"
		case "pagination":
			token := "next-page"
			response.ContinuationToken = &token
		case "not found":
			sendJSONError(w, "private-value", http.StatusNotFound)
			return
		case "unauthorized":
			sendJSONError(w, "private-value", http.StatusUnauthorized)
			return
		case "malformed":
			fmt.Fprint(w, `{"data": "private-value"`)
			return
		case "empty body":
			return
		}
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}).Methods("POST")

	transport := httpmock.NewMockTransport()
	transport.RegisterNoResponder(func(req *http.Request) (*http.Response, error) {
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		return resp.Result(), nil
	})
	client := embedded.NewSecretsManagerClient("http://127.0.0.1", "test-device", "test",
		embedded.WithSecretsManagerHttpOptions(webapi.WithCustomClient(http.Client{Transport: transport})),
	)
	org, err := backend.ClientCreateNewOrganization()
	require.NoError(t, err)
	token, err := backend.ClientCreateAccessToken(org)
	require.NoError(t, err)
	require.NoError(t, client.LoginWithAccessToken(t.Context(), token))
	project, err := client.CreateProject(t.Context(), models.Project{Name: "test-project", OrganizationID: org})
	require.NoError(t, err)
	var ids []string
	wantSecrets := map[string]batchSecretModel{}
	for i := range secretCount {
		name := fmt.Sprintf("secret-%02d", i)
		value := fmt.Sprintf("private-value-%d\"\n雪", i)
		// Empty secret values remain valid provider data.
		if i == 0 {
			value = ""
		}
		secret, err := client.CreateSecret(t.Context(), models.Secret{
			Key:            name,
			Value:          value,
			Note:           "private-note",
			ProjectID:      project.ID,
			OrganizationID: org,
		})
		require.NoError(t, err)
		ids = append(ids, secret.ID)
		wantSecrets[secret.ID] = batchSecretModel{Key: name, Value: value, Note: "private-note"}
	}
	ids = append(ids, strings.ToUpper(ids[1]))
	server := ephemeralSecretServer(t, &ProviderClients{SecretsManager: client})

	tests := []string{
		"success",
		"missing",
		"duplicate",
		"unexpected",
		"corrupt ciphertext",
		"pagination",
		"not found",
		"unauthorized",
		"malformed",
		"empty body",
	}

	for _, result := range tests {
		t.Run(result, func(t *testing.T) {
			outcome = result
			before := batchCalls
			resp, err := server.OpenEphemeralResource(t.Context(), &tfprotov6.OpenEphemeralResourceRequest{
				TypeName: "bitwarden_secrets",
				Config:   ephemeralSecretsConfig(t, ids, nil),
			})
			require.NoError(t, err)
			assert.Equal(t, before+1, batchCalls)
			if result == "success" {
				require.Empty(t, resp.Diagnostics)
				assert.Equal(t, wantSecrets, ephemeralSecretsRecords(t, resp))
			} else {
				require.NotEmpty(t, resp.Diagnostics)
				assertEphemeralSecretsNoValues(t, resp)
			}
			for _, marker := range []string{"private-value", "private-note"} {
				assert.NotContains(t, fmt.Sprint(resp.Diagnostics), marker)
			}
		})
	}

	before := transport.GetTotalCallCount()
	secrets, err := client.GetSecretsByIDs(t.Context(), nil)
	require.NoError(t, err)
	assert.Empty(t, secrets)
	assert.Equal(t, before, transport.GetTotalCallCount())
}
