package provider

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

var (
	_ ephemeral.EphemeralResource              = &secretEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &secretEphemeralResource{}
)

type secretEphemeralResource struct {
	clients *ProviderClients
}

type secretEphemeralResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	Note           types.String `tfsdk:"note"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
}

func NewSecretEphemeralResource() ephemeral.EphemeralResource {
	return &secretEphemeralResource{}
}

func (r *secretEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema_definition.SecretEphemeralResourceSchema()
}

func (r *secretEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	clients, ok := clientsFromProviderData(req.ProviderData, &resp.Diagnostics)
	if !ok {
		return
	}

	r.clients = clients
}

func (r *secretEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data secretEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := requireSecretsManager(r.clients, &resp.Diagnostics)
	if !ok {
		return
	}

	var secret *models.Secret
	var err error

	// The CLI logs stdout and stderr before parsing its response. Either can
	// contain decrypted secrets, including unrelated secrets from a key lookup.
	// Scope the filter to this read so command status logs remain available.
	ctx = tflog.OmitLogWithFieldKeys(ctx, "stdout", "stderr")

	if !data.ID.IsNull() {
		secret, err = client.GetSecret(ctx, models.Secret{ID: data.ID.ValueString()})
	} else {
		secret, err = client.GetSecretByKey(ctx, data.Key.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to read secret", ephemeralSecretError(err))
		return
	}

	data.ID = types.StringValue(secret.ID)
	data.Key = types.StringValue(secret.Key)
	data.Value = types.StringValue(secret.Value)
	data.Note = types.StringValue(secret.Note)
	data.OrganizationID = types.StringValue(secret.OrganizationID)
	data.ProjectID = types.StringValue(secret.ProjectID)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func ephemeralSecretError(err error) string {
	// Client errors can include raw CLI output containing decrypted secrets.
	// Only known errors are safe to include in Terraform diagnostics.
	for _, known := range []error{
		models.ErrObjectNotFound,
		models.ErrNoObjectFoundMatchingFilter,
		models.ErrTooManyObjectsFound,
		models.ErrLoggedOut,
		context.Canceled,
		context.DeadlineExceeded,
	} {
		if errors.Is(err, known) {
			return known.Error()
		}
	}

	if httpErr, ok := webapi.IsHTTPError(err); ok {
		detail := fmt.Sprintf("Bitwarden server returned HTTP %d (%s).", httpErr.StatusCode, http.StatusText(httpErr.StatusCode))
		switch httpErr.StatusCode {
		case http.StatusUnauthorized:
			detail += " Check the provider credentials."
		case http.StatusForbidden:
			detail += " Check the access token's permissions for the requested secrets."
		case http.StatusTooManyRequests:
			detail += " The request was rate limited; retry later."
		}
		return detail + " Server response details are omitted because they may contain secret values."
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "Unable to resolve the Bitwarden server (DNS error). Check the server address and DNS connectivity."
	}

	var certificateErr *tls.CertificateVerificationError
	if errors.As(err, &certificateErr) {
		return "Unable to verify the Bitwarden server's TLS certificate. Check the server certificate and trusted CA configuration."
	}

	var networkErr net.Error
	if errors.As(err, &networkErr) {
		if networkErr.Timeout() {
			return "The Bitwarden request timed out. Check server connectivity and retry."
		}
		return "A network error occurred while contacting Bitwarden. Check server connectivity."
	}

	return "Bitwarden could not read the secret. Check the provider credentials, access permissions, and server connectivity. Client error details are omitted because they may contain secret values."
}
