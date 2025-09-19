package project

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var ProjectResourceDataV0 = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"key":        tftypes.String,
		"version":    tftypes.Number,
		"name":       tftypes.String,
		"currencies": tftypes.List{ElementType: tftypes.String},
		"countries":  tftypes.List{ElementType: tftypes.String},
		"languages":  tftypes.List{ElementType: tftypes.String},

		"enable_search_index_products": tftypes.Bool,
		"enable_search_index_orders":   tftypes.Bool,

		"carts": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"country_tax_rate_fallback_enabled":   tftypes.Bool,
				"delete_days_after_last_modification": tftypes.Number,
			},
		},
		"messages": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"enabled":                    tftypes.Bool,
				"delete_days_after_creation": tftypes.Number,
			},
		},
		"external_oauth": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"url":                  tftypes.String,
				"authorization_header": tftypes.String,
			},
		},
		"shipping_rate_input_type": tftypes.String,
		"shipping_rate_cart_classification_value": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key": tftypes.String,
					"label": tftypes.Map{
						ElementType: tftypes.String,
					},
				},
			},
		},
	},
}

var ProjectsCartDataV1 = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"country_tax_rate_fallback_enabled":   tftypes.Bool,
		"delete_days_after_last_modification": tftypes.Number,
		"price_rounding_mode":                 tftypes.String,
		"tax_rounding_mode":                   tftypes.String,
	},
}

var ProjectResourceDataV1 = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"key":        tftypes.String,
		"version":    tftypes.Number,
		"name":       tftypes.String,
		"currencies": tftypes.List{ElementType: tftypes.String},
		"countries":  tftypes.List{ElementType: tftypes.String},
		"languages":  tftypes.List{ElementType: tftypes.String},

		"enable_search_index_products":       tftypes.Bool,
		"enable_search_index_product_search": tftypes.Bool,
		"enable_search_index_orders":         tftypes.Bool,
		"enable_search_index_customers":      tftypes.Bool,
		"enable_search_index_business_units": tftypes.Bool,

		"carts": tftypes.List{
			ElementType: ProjectsCartDataV1,
		},
		"shopping_lists": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"delete_days_after_last_modification": tftypes.Number,
				},
			},
		},
		"messages": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"enabled":                    tftypes.Bool,
					"delete_days_after_creation": tftypes.Number,
				},
			},
		},
		"external_oauth": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"url":                  tftypes.String,
					"authorization_header": tftypes.String,
				},
			},
		},
		"shipping_rate_input_type": tftypes.String,
		"shipping_rate_cart_classification_value": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"key": tftypes.String,
					"label": tftypes.Map{
						ElementType: tftypes.String,
					},
				},
			},
		},
		"business_units": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"my_business_unit_status_on_creation":             tftypes.String,
					"my_business_unit_associate_role_key_on_creation": tftypes.String,
				},
			},
		},
	},
}

// Move from version 0 to current. Version 1 changed some items from single
// blocks to lists with a max of 1. This was needed since sdk v2 did only
// support that approach.
// Moved from v0 to v1 in v1.0.0.pre0, see https://github.com/labd/terraform-provider-commercetools/pull/196
func upgradeStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	rawStateValue, err := req.RawState.Unmarshal(ProjectResourceDataV0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Unmarshal Prior State",
			err.Error(),
		)
		return
	}

	var rawState map[string]tftypes.Value
	if err := rawStateValue.As(&rawState); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Convert Prior State",
			err.Error(),
		)
		return
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(
		ProjectResourceDataV1,
		tftypes.NewValue(ProjectResourceDataV1, map[string]tftypes.Value{
			"id":         rawState["id"],
			"key":        rawState["key"],
			"version":    rawState["version"],
			"name":       rawState["name"],
			"currencies": rawState["currencies"],
			"countries":  rawState["countries"],
			"languages":  rawState["languages"],

			"carts":          tftypes.NewValue(ProjectResourceDataV1.AttributeTypes["carts"], []tftypes.Value{transformCarts(rawState["carts"])}),
			"shopping_lists": valueToList(rawState, "shopping_lists"),
			"messages":       valueToList(rawState, "messages"),
			"external_oauth": valueToList(rawState, "external_oauth"),

			"shipping_rate_input_type":                rawState["shipping_rate_input_type"],
			"shipping_rate_cart_classification_value": rawState["shipping_rate_cart_classification_value"],

			// Values that didn't exist yet
			"enable_search_index_products":       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			"enable_search_index_product_search": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			"enable_search_index_orders":         tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			"enable_search_index_customers":      tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			"enable_search_index_business_units": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			"business_units":                     valueToList(nil, "business_units"),
		}),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Convert Upgraded State",
			err.Error(),
		)
		return
	}

	resp.DynamicValue = &dynamicValue
}

func transformCarts(state tftypes.Value) tftypes.Value {
	countryTaxRateFallbackEnabled, err := state.ApplyTerraform5AttributePathStep(tftypes.AttributeName("country_tax_rate_fallback_enabled"))
	if err != nil {
		panic(err)
	}
	deleteDaysAfterLastModification, err := state.ApplyTerraform5AttributePathStep(tftypes.AttributeName("delete_days_after_last_modification"))
	if err != nil {
		panic(err)
	}

	priceRoundingMode, err := state.ApplyTerraform5AttributePathStep(tftypes.AttributeName("price_rounding_mode"))
	if err != nil {
		if errors.Is(err, tftypes.ErrInvalidStep) {
			priceRoundingMode = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		} else {
			panic(err)
		}
	}

	taxRoundingMode, err := state.ApplyTerraform5AttributePathStep(tftypes.AttributeName("tax_rounding_mode"))
	if err != nil {
		if errors.Is(err, tftypes.ErrInvalidStep) {
			taxRoundingMode = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		} else {
			panic(err)
		}
	}

	nv := map[string]tftypes.Value{
		"country_tax_rate_fallback_enabled":   countryTaxRateFallbackEnabled.(tftypes.Value),
		"delete_days_after_last_modification": deleteDaysAfterLastModification.(tftypes.Value),
		"price_rounding_mode":                 priceRoundingMode.(tftypes.Value),
		"tax_rounding_mode":                   taxRoundingMode.(tftypes.Value),
	}

	return tftypes.NewValue(ProjectsCartDataV1, nv)
}

func valueToList(state map[string]tftypes.Value, key string) tftypes.Value {
	if state[key].IsNull() {
		return tftypes.NewValue(
			ProjectResourceDataV1.AttributeTypes[key],
			[]tftypes.Value{},
		)
	}

	if state[key].IsKnown() {
		return tftypes.NewValue(
			ProjectResourceDataV1.AttributeTypes[key],
			[]tftypes.Value{state[key]},
		)
	}
	return state[key]
}
