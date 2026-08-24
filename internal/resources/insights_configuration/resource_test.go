package insights_configuration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestResourceSchema validates the schema against the framework's own rules,
// catching mistakes such as attributes that are neither optional, required nor
// computed, or invalid nested block configurations.
func TestResourceSchema(t *testing.T) {
	ctx := context.Background()
	resp := &fwresource.SchemaResponse{}
	NewResource().Schema(ctx, fwresource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned diagnostics: %+v", resp.Diagnostics)
	}

	if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %+v", diags)
	}
}

// TestModelMatchesSchema converts a fully populated model into the schema's
// type. This catches tfsdk struct tags that do not line up with the schema,
// which otherwise only surfaces as an error during an apply.
func TestModelMatchesSchema(t *testing.T) {
	ctx := context.Background()
	resp := &fwresource.SchemaResponse{}
	NewResource().Schema(ctx, fwresource.SchemaRequest{}, resp)

	model := InsightsConfiguration{
		ID:             types.StringValue("my-project"),
		ProjectKey:     types.StringValue("my-project"),
		Active:         types.BoolValue(true),
		LastModifiedAt: types.StringValue("2024-12-12T12:01:03Z"),
		Providers: []Provider{
			{
				Type:             types.StringValue(DynatraceActiveGate),
				Active:           types.BoolValue(true),
				Valid:            types.BoolValue(true),
				EventTypes:       []types.String{types.StringValue("Metrics")},
				Region:           types.StringNull(),
				Site:             types.StringNull(),
				EnvironmentID:    types.StringValue("dtMyEnvironment"),
				ActiveGateDomain: types.StringValue("my-active-gate.example.com"),
				ActiveGatePort:   types.Int64Value(9999),
				Endpoint:         types.StringNull(),
				ApiKey:           types.StringValue("dtabc1234"),
				Headers:          []Header{},
			},
			{
				Type:       types.StringValue(Otlp),
				Active:     types.BoolValue(true),
				Valid:      types.BoolNull(),
				EventTypes: []types.String{types.StringValue("Logs")},
				Endpoint:   types.StringValue("https://myotlpprovider.com:14317"),
				Headers: []Header{{
					Name:  types.StringValue("api-key"),
					Value: types.StringValue("apiKey1234"),
				}},
			},
		},
		AdditionalAttributes: []Attribute{{
			EventType: types.StringValue("Metrics"),
			Name:      types.StringValue("team"),
			Value:     types.StringValue("commercetools"),
		}},
	}

	var target attr.Value
	if diags := tfsdk.ValueFrom(ctx, model, resp.Schema.Type(), &target); diags.HasError() {
		t.Fatalf("model does not match schema: %+v", diags)
	}
}
