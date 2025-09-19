package project

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/models"
)

func Test_upgradeStateV0(t *testing.T) {
	oldState := []byte(`
	  {
		"carts":{
			"country_tax_rate_fallback_enabled":false,
			"delete_days_after_last_modification":10
		},
		"countries":[],
		"currencies":[
			"EUR"
		],
		"external_oauth":null,
		"id":"my-project",
		"key":"my-project",
		"languages":[
			"nl"
		],
		"messages":{
			"delete_days_after_creation":15,
			"enabled":false
		},
		"name":"My Project",
		"shipping_rate_cart_classification_value":[
			{
			"key":"Small",
			"label":{
				"en":"Small",
				"nl":"Klein"
			}
			}
		],
		"shipping_rate_input_type":"CartClassification",
		"version":180
	  }
	`)

	expected := Project{
		Version: types.Int64Value(180),
		ID:      types.StringValue("my-project"),
		Key:     types.StringValue("my-project"),
		Name:    types.StringValue("My Project"),

		Currencies: []types.String{types.StringValue("EUR")},
		Countries:  []types.String{},
		Languages:  []types.String{types.StringValue("nl")},

		EnableSearchIndexProducts:      types.BoolUnknown(),
		EnableSearchIndexProductSearch: types.BoolUnknown(),
		EnableSearchIndexOrders:        types.BoolUnknown(),
		EnableSearchIndexCustomers:     types.BoolUnknown(),
		EnableSearchIndexBusinessUnits: types.BoolUnknown(),
		BusinessUnits:                  []BusinessUnits{},

		ExternalOAuth: []ExternalOAuth{},
		Carts: []Carts{
			{
				CountryTaxRateFallbackEnabled:   types.BoolValue(false),
				DeleteDaysAfterLastModification: types.Int64Value(10),
				PriceRoundingMode:               types.StringUnknown(),
				TaxRoundingMode:                 types.StringUnknown(),
			},
		},
		ShoppingLists: []ShoppingList{},
		Messages: []Messages{
			{
				Enabled:                 types.BoolValue(false),
				DeleteDaysAfterCreation: types.Int64Value(15),
			},
		},
		ShippingRateInputType: types.StringValue("CartClassification"),
		ShippingRateCartClassificationValue: []models.CustomFieldLocalizedEnumValue{
			{
				Key: types.StringValue("Small"),
				Label: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en": types.StringValue("Small"),
					"nl": types.StringValue("Klein"),
				}),
			},
		},
	}

	ctx := context.Background()
	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: oldState,
		},
	}
	resp := resource.UpgradeStateResponse{}
	upgradeStateV0(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
	require.NotNil(t, resp.DynamicValue)

	// Create the state based on the current schema
	s := getCurrentSchema()
	upgradedStateValue, err := resp.DynamicValue.Unmarshal(s.Type().TerraformType(ctx))
	require.NoError(t, err)
	state := tfsdk.State{
		Raw:    upgradedStateValue,
		Schema: s,
	}

	res := Project{}
	diags := state.Get(ctx, &res)
	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, expected, res)
}

func getCurrentSchema() schema.Schema {
	ctx := context.Background()
	res := NewResource()

	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	res.Schema(ctx, req, &resp)
	return resp.Schema
}
