package main

import (
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
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

	providerAddr string = "registry.terraform.io/maxlaverse/bitwarden"
)

func main() {
	_ = commit

	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	serverFactory, err := provider.NewProviderServer(version)
	if err != nil {
		log.Fatal(err.Error())
	}

	var serveOpts []tf6server.ServeOpt
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err := tf6server.Serve(providerAddr, serverFactory, serveOpts...); err != nil {
		log.Fatal(err.Error())
	}
}
