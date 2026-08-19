package state

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
)

func TestState_UpdateActions(t *testing.T) {
	cases := []struct {
		name     string
		state    State
		plan     State
		expected platform.StateUpdate
	}{
		{
			"no changes when initial is unset in both state and plan",
			State{Initial: types.BoolNull()},
			State{Initial: types.BoolNull()},
			platform.StateUpdate{
				Actions: []platform.StateUpdateAction{},
			},
		},
		{
			"no changes when initial is false in both state and plan",
			State{Initial: types.BoolValue(false)},
			State{Initial: types.BoolValue(false)},
			platform.StateUpdate{
				Actions: []platform.StateUpdateAction{},
			},
		},
		{
			"no changes when initial is unset in state and false in plan",
			State{Initial: types.BoolNull()},
			State{Initial: types.BoolValue(false)},
			platform.StateUpdate{
				Actions: []platform.StateUpdateAction{},
			},
		},
		{
			"change initial when unset in state and true in plan",
			State{Initial: types.BoolNull()},
			State{Initial: types.BoolValue(true)},
			platform.StateUpdate{
				Actions: []platform.StateUpdateAction{
					platform.StateChangeInitialAction{Initial: true},
				},
			},
		},
		{
			"change initial when true in state and false in plan",
			State{Initial: types.BoolValue(true)},
			State{Initial: types.BoolValue(false)},
			platform.StateUpdate{
				Actions: []platform.StateUpdateAction{
					platform.StateChangeInitialAction{Initial: false},
				},
			},
		},
		{
			"update the name and keep initial untouched",
			State{
				Initial: types.BoolNull(),
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en": types.StringValue("Payment failure"),
				}),
			},
			State{
				Initial: types.BoolNull(),
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en": types.StringValue("Payment failed"),
				}),
			},
			platform.StateUpdate{
				Actions: []platform.StateUpdateAction{
					platform.StateSetNameAction{
						Name: platform.LocalizedString{"en": "Payment failed"},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := c.state.updateActions(c.plan)
			assert.EqualValues(t, c.expected, result)
		})
	}
}
