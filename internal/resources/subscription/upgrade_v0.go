package subscription

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Upgrade from V0 to V1
func upgradeStateV0(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
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
			"destination": valueDestinationV1(rawState, "destination"),
			"format":      valueToFormatV1(rawState, "format"),
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
