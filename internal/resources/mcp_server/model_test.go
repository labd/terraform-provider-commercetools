package mcp_server

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceSchema(t *testing.T) {
	ctx := context.Background()
	resp := &fwresource.SchemaResponse{}
	NewResource().Schema(ctx, fwresource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	diags := resp.Schema.ValidateImplementation(ctx)
	require.False(t, diags.HasError(), diags)
}

func TestParameterJSONKeepsTheConfiguredSpelling(t *testing.T) {
	prior := []ToolCustomization{{
		Tool: types.StringValue("read_products"),
		Parameters: []ParameterOverride{{
			Name:        types.StringValue("where"),
			DefaultJSON: types.StringValue(`{ "b": 1, "a": [1, 2] }`),
		}},
	}}
	read := fromCustomizations([]platform.ToolCustomization{{
		Tool: "read_products",
		Parameters: []platform.ParameterOverride{{
			Name:    "where",
			Default: map[string]any{"a": []any{1.0, 2.0}, "b": 1.0},
		}},
	}}, prior)
	assert.Equal(t, `{ "b": 1, "a": [1, 2] }`, read[0].Parameters[0].DefaultJSON.ValueString())
	assert.True(t, read[0].Parameters[0].ConstJSON.IsNull())
}

func TestUpdateActionsIgnoreToolOrder(t *testing.T) {
	state := McpServer{
		Version:            types.Int64Value(3),
		AuthenticationMode: types.StringValue("ClientCredentials"),
		State:              types.StringValue("Disabled"),
		Tools:              []types.String{types.StringValue("read_products"), types.StringValue("read_orders")},
	}
	plan := state
	plan.Tools = []types.String{types.StringValue("read_orders"), types.StringValue("read_products")}
	update, err := state.updateActions(plan)
	require.NoError(t, err)
	assert.Empty(t, update.Actions)

	plan.State = types.StringValue("Enabled")
	plan.Tools = append(plan.Tools, types.StringValue("read_carts"))
	update, err = state.updateActions(plan)
	require.NoError(t, err)
	assert.Equal(t, []platform.McpServerUpdateAction{
		platform.McpServerSetStateAction{State: platform.McpServerStateEnabled},
		platform.McpServerSetToolsAction{Tools: []platform.McpServerTool{"read_orders", "read_products", "read_carts"}},
	}, update.Actions)
}
