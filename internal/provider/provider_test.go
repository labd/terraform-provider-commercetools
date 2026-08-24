package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Smoke test: the provider server must expose the new resource in its schema.
func TestProviderExposesInsightsConfiguration(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(New("test"))()

	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema: %v", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("schema diagnostic: %s: %s", d.Summary, d.Detail)
		}
	}

	schema, ok := resp.ResourceSchemas["commercetools_insights_configuration"]
	if !ok {
		t.Fatal("commercetools_insights_configuration not registered")
	}
	t.Logf("registered with %d attributes and %d blocks",
		len(schema.Block.Attributes), len(schema.Block.BlockTypes))
}
