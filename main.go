package main

import (
	"flag"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/labd/terraform-provider-commercetools/commercetools"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"
	commit  string = "snapshot"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	fullVersion := fmt.Sprintf("%s (%s)", version, commit)

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: commercetools.New(fullVersion),
		Debug:        debugMode,
	})
}
