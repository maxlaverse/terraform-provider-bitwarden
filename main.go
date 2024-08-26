package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/provider"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "0.0.1"

	// goreleaser can also pass the specific commit if you want
	commit string = ""

	providerAddr string = "registry.terraform.io/maxlaverse/terraform-provider-bitwarden"
)

func main() {
	opts := &plugin.ServeOpts{ProviderFunc: provider.New(version)}

	plugin.Serve(opts)
}
