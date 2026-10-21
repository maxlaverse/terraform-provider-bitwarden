//go:build offline

package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/require"
)

func TestEphemeralSecretEmbedded(t *testing.T) {
	backend := NewTestSecretsManager()
	router := mux.NewRouter()
	router.HandleFunc("/identity/connect/token", backend.handlerLogin).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/projects", backend.handlerCreateProject).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/secrets", backend.handlerCreateGetSecret).Methods("POST", "GET")
	router.HandleFunc("/api/secrets/{secretId}", backend.handlerGetSecret).Methods("GET")
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
	wanted := models.Secret{Key: "duplicate-key", Value: "private-value", Note: "private-note", ProjectID: project.ID, OrganizationID: org}
	secret, err := client.CreateSecret(t.Context(), wanted)
	require.NoError(t, err)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ephemeralCredentialFactories(&ProviderClients{SecretsManager: client}, *secret),
		Steps: []resource.TestStep{
			{
				Config: ephemeralSecretConsumerConfig(fmt.Sprintf("id = %q", secret.ID)),
				Check:  resource.TestCheckOutput("project_name", "public-project"),
			},
			{
				Config: ephemeralSecretConsumerConfig(fmt.Sprintf("key = %q", wanted.Key)),
				Check:  resource.TestCheckOutput("project_name", "public-project"),
			},
			{
				Config:      ephemeralSecretConsumerConfig(`id = "missing-secret"`),
				ExpectError: regexp.MustCompile("object not found"),
			},
			{
				Config:      ephemeralSecretConsumerConfig(`key = "missing-key"`),
				ExpectError: regexp.MustCompile("object not found"),
			},
			{
				PreConfig: func() {
					_, err := client.CreateSecret(t.Context(), wanted)
					require.NoError(t, err)
				},
				Config:      ephemeralSecretConsumerConfig(fmt.Sprintf("key = %q", wanted.Key)),
				ExpectError: regexp.MustCompile("too many objects found"),
			},
			{
				Config: ephemeralSecretConsumerConfig(fmt.Sprintf("id = %q", secret.ID)),
				Check:  resource.TestCheckOutput("project_name", "public-project"),
			},
		},
	})
}
