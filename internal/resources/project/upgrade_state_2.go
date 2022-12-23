package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var ProjectResourceDataV1 = tftypes.Object{
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

		"carts": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"country_tax_rate_fallback_enabled":   tftypes.String,
					"delete_days_after_last_modification": tftypes.String,
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
	},
}

var ProjectResourceDataV2 = tftypes.Object{
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
				"country_tax_rate_fallback_enabled":   tftypes.String,
				"delete_days_after_last_modification": tftypes.String,
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

// Schema version 0 is fully compatible with current version
func upgradeStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
}

// Move from version 1 to current. Changes where needed when we upgrade to
// sdk v2 which always uses lists for nested blocks. So carts = [{..}] instead of
// carts = {..}
func upgradeStateV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	rawStateValue, err := req.RawState.Unmarshal(ProjectResourceDataV1)
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

	carts, diags := ItemFromList(rawState, "carts")
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	messages, diags := ItemFromList(rawState, "messages")
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	external_oauth, diags := ItemFromList(rawState, "external_oauth")
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(
		ProjectResourceDataV2,
		tftypes.NewValue(ProjectResourceDataV2, map[string]tftypes.Value{
			"id":                           rawState["id"],
			"key":                          rawState["key"],
			"version":                      rawState["version"],
			"name":                         rawState["name"],
			"currencies":                   rawState["currencies"],
			"countries":                    rawState["countries"],
			"languages":                    rawState["languages"],
			"enable_search_index_products": rawState["enable_search_index_products"],
			"enable_search_index_orders":   rawState["enable_search_index_orders"],

			"carts":          carts,
			"messages":       messages,
			"external_oauth": external_oauth,

			"shipping_rate_input_type":                rawState["shipping_rate_input_type"],
			"shipping_rate_cart_classification_value": rawState["shipping_rate_cart_classification_value"],
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

func ItemFromList(rawState map[string]tftypes.Value, key string) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !rawState[key].IsNull() {
		val := []tftypes.Value{}
		if err := rawState[key].As(&val); err != nil {
			diags.AddAttributeError(
				path.Root(key),
				fmt.Sprintf("Unable to Convert Prior State (%s)", key),
				err.Error(),
			)
			return tftypes.Value{}, diags
		}
		if len(val) > 0 {
			result := map[string]tftypes.Value{}
			if err := val[0].As(&result); err != nil {
				diags.AddAttributeError(
					path.Root(key),
					fmt.Sprintf("Unable to Convert Prior State (%s)", key),
					err.Error(),
				)
				return tftypes.Value{}, diags
			}
			value := tftypes.NewValue(ProjectResourceDataV2.AttributeTypes[key], result)
			return value, diags
		}
		value := tftypes.NewValue(ProjectResourceDataV2.AttributeTypes[key], nil)
		return value, diags
	}
	return tftypes.Value{}, diags
}
