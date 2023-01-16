package subscription

import (
	"context"

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
// Schema version 2 moves us to Single nested blocks, but it turned out to be
// not working correctly in terraform for now. So we moved back to the v1
// approach
func upgradeStateV2(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	rawStateValue, err := req.RawState.Unmarshal(SubscriptionResourceV2)
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
		SubscriptionResourceV1,
		tftypes.NewValue(SubscriptionResourceV1, map[string]tftypes.Value{
			"id":          rawState["id"],
			"key":         rawState["key"],
			"version":     rawState["version"],
			"changes":     rawState["changes"],
			"destination": valueToList(rawState, "destination"),
			"format":      valueToList(rawState, "format"),
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

func valueToList(state map[string]tftypes.Value, key string) tftypes.Value {
	if state[key].IsNull() {
		return tftypes.NewValue(
			SubscriptionResourceV1.AttributeTypes[key],
			[]tftypes.Value{},
		)
	}

	if state[key].IsKnown() {
		return tftypes.NewValue(
			SubscriptionResourceV1.AttributeTypes[key],
			[]tftypes.Value{state[key]},
		)
	}
	return state[key]
}
