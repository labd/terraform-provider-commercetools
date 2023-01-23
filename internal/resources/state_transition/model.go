package state_transition

import (
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type StateTransition struct {
	ID      types.String   `tfsdk:"id"`
	From    types.String   `tfsdk:"from"`
	To      []types.String `tfsdk:"to"`
	Version types.Int64    `tfsdk:"-"`
}

func (s StateTransition) updateActions(plan StateTransition) platform.StateUpdate {
	result := platform.StateUpdate{
		Version: int(s.Version.ValueInt64()),
		Actions: []platform.StateUpdateAction{},
	}

	if !reflect.DeepEqual(s.To, plan.To) {
		result.Actions = append(
			result.Actions,
			platform.StateSetTransitionsAction{
				Transitions: pie.Map(plan.To, func(v types.String) platform.StateResourceIdentifier {
					return platform.StateResourceIdentifier{
						ID: utils.StringRef(v.ValueString()),
					}
				}),
			})
	}

	return result
}

func NewStateTransitionFromNative(n *platform.State) StateTransition {
	return StateTransition{
		ID:   types.StringValue(n.ID),
		From: types.StringValue(n.ID),
		To: pie.Map(n.Transitions, func(ref platform.StateReference) types.String {
			return types.StringValue(ref.ID)
		}),
	}
}
