package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/labd/terraform-provider-commercetools/commercetools"
)

func main() {
	tfsdk.Serve(context.Background(), commercetools.New, tfsdk.ServeOpts{
		Name: "commercetools",
	})
}
