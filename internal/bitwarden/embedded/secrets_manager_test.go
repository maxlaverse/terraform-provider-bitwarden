package embedded

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded/fixtures"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

func newMockedSecretsManager(client webapi.Client) (secretsManager, func()) {
	httpmock.Activate()

	return secretsManager{
		serverURL: fixtures.ServerURL,
		client:    client,
	}, httpmock.DeactivateAndReset
}
