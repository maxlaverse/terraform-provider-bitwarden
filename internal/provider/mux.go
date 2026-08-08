package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

// NewProviderServer returns a Protocol 6 provider server that muxes the
// Plugin Framework shell (New) with the SDKv2 implementation (NewSDK).
// Resources/data sources move to Framework one at a time (folder, project, org
// data sources, …); each is removed from NewSDK's maps as it migrates so type
// names never collide.
func NewProviderServer(version string) (func() tfprotov6.ProviderServer, error) {
	ctx := context.Background()

	upgradedSDK, err := tf5to6server.UpgradeServer(ctx, NewSDK(version)().GRPCProvider)
	if err != nil {
		return nil, fmt.Errorf("upgrade SDKv2 provider to protocol 6: %w", err)
	}

	servers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(New(version)()),
		func() tfprotov6.ProviderServer { return upgradedSDK },
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, servers...)
	if err != nil {
		return nil, fmt.Errorf("mux Framework and SDKv2 providers: %w", err)
	}

	return muxServer.ProviderServer, nil
}
