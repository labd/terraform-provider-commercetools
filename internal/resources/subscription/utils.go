package subscription

import "github.com/hashicorp/terraform-plugin-go/tftypes"

// SubscriptionResourceV1 represents the currently used structure of the
// subscription resource. This is used to map legacy structures to the current
// required structure.
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
			ElementType: destinationType,
		},
		"format": tftypes.List{
			ElementType: formatType,
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

var formatType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"type":                 tftypes.String,
		"cloud_events_version": tftypes.String,
	},
}

var destinationType = tftypes.Object{
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
		"bootstrap_server":  tftypes.String,
		"api_key":           tftypes.String,
		"api_secret":        tftypes.String,
		"acks":              tftypes.String,
		"key":               tftypes.String,
	},
}

func valueToFormatV1(state map[string]tftypes.Value, key string) tftypes.Value {
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

func valueDestinationV1(state map[string]tftypes.Value, key string) tftypes.Value {
	if state[key].IsNull() {
		return tftypes.NewValue(
			SubscriptionResourceV1.AttributeTypes[key],
			[]tftypes.Value{},
		)
	}

	if state[key].IsKnown() {
		newVal := map[string]tftypes.Value{}
		val := state[key]
		err := val.As(&newVal)
		if err != nil {
			panic(err)
		}

		//Add additional fields to make the older versions compatible with new fields
		newVal["bootstrap_server"] = tftypes.NewValue(tftypes.String, nil)
		newVal["api_key"] = tftypes.NewValue(tftypes.String, nil)
		newVal["api_secret"] = tftypes.NewValue(tftypes.String, nil)
		newVal["acks"] = tftypes.NewValue(tftypes.String, nil)
		newVal["key"] = tftypes.NewValue(tftypes.String, nil)

		val = tftypes.NewValue(destinationType, newVal)

		return tftypes.NewValue(
			SubscriptionResourceV1.AttributeTypes[key],
			[]tftypes.Value{val},
		)
	}
	return state[key]
}
