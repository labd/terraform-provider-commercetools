package subscription

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var SubscriptionResourceV1 = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":      tftypes.String,
		"key":     tftypes.String,
		"version": tftypes.Number,

		"changes": tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"resource_type_ids": tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
		},
		"destination": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"type":              tftypes.String,
					"topic_arn":         tftypes.String,
					"queue_url":         tftypes.String,
					"region":            tftypes.String,
					"account_id":        tftypes.String,
					"access_key":        tftypes.String,
					"access_secret":     tftypes.String,
					"uri":               tftypes.String,
					"connection_string": tftypes.String,
					"project_id":        tftypes.String,
					"topic":             tftypes.String,
				},
			},
		},
		"format": tftypes.List{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"type":                 tftypes.String,
					"cloud_events_version": tftypes.String,
				},
			},
		},
		"message": tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"resource_type_id": tftypes.String,
					"types": tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
		},
	},
}

var SubscriptionResourceV2 = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":      tftypes.String,
		"key":     tftypes.String,
		"version": tftypes.Number,

		"changes": tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"resource_type_ids": tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
		},
		"destination": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"type":              tftypes.String,
				"topic_arn":         tftypes.String,
				"queue_url":         tftypes.String,
				"region":            tftypes.String,
				"account_id":        tftypes.String,
				"access_key":        tftypes.String,
				"access_secret":     tftypes.String,
				"uri":               tftypes.String,
				"connection_string": tftypes.String,
				"project_id":        tftypes.String,
				"topic":             tftypes.String,
			},
		},
		"format": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"type":                 tftypes.String,
				"cloud_events_version": tftypes.String,
			},
		},
		"message": tftypes.Set{
			ElementType: tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"resource_type_id": tftypes.String,
					"types": tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
		},
	},
}

// Schema version 1 used a list for destination and format since
// that single nested blocks were not supported in sdk v2 (it was in sdk v1)
func upgradeStateV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	rawStateValue, err := req.RawState.Unmarshal(SubscriptionResourceV1)
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

	destination, diags := ItemFromList(rawState, "destination")
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	format, diags := ItemFromList(rawState, "format")
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(
		SubscriptionResourceV2,
		tftypes.NewValue(SubscriptionResourceV2, map[string]tftypes.Value{
			"id":          rawState["id"],
			"key":         rawState["key"],
			"version":     rawState["version"],
			"changes":     rawState["changes"],
			"destination": destination,
			"format":      format,
			"message":     rawState["message"],
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
			value := tftypes.NewValue(SubscriptionResourceV2.AttributeTypes[key], result)
			return value, diags
		}
		value := tftypes.NewValue(SubscriptionResourceV2.AttributeTypes[key], nil)
		return value, diags
	}
	return tftypes.Value{}, diags
}
