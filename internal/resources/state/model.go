package state

import (
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type State struct {
	ID          types.String                     `tfsdk:"id"`
	Key         types.String                     `tfsdk:"key"`
	Version     types.Int64                      `tfsdk:"version"`
	Type        types.String                     `tfsdk:"type"`
	Name        customtypes.LocalizedStringValue `tfsdk:"name"`
	Description customtypes.LocalizedStringValue `tfsdk:"description"`
	Initial     types.Bool                       `tfsdk:"initial"`
	Roles       []types.String                   `tfsdk:"roles"`
}

func NewStateFromNative(n *platform.State) State {
	res := State{
		ID:          types.StringValue(n.ID),
		Version:     types.Int64Value(int64(n.Version)),
		Key:         types.StringValue(n.Key),
		Name:        utils.FromOptionalLocalizedString(n.Name),
		Description: utils.FromOptionalLocalizedString(n.Description),
		Type:        types.StringValue(string(n.Type)),
		Initial:     types.BoolValue(n.Initial),
	}

	// If the roles is empty we want to keep the value as null and not an empty
	// list
	if len(n.Roles) > 0 {
		res.Roles = pie.Map(n.Roles, func(v platform.StateRoleEnum) types.String {
			return types.StringValue(string(v))
		})
	}
	return res
}

func (s State) draft() platform.StateDraft {
	result := platform.StateDraft{
		Key:         s.Key.ValueString(),
		Type:        platform.StateTypeEnum(s.Type.ValueString()),
		Name:        s.Name.ValueLocalizedStringRef(),
		Description: s.Description.ValueLocalizedStringRef(),
		Initial:     utils.BoolRef(s.Initial.ValueBool()),
		Roles: pie.Map(s.Roles, func(v types.String) platform.StateRoleEnum {
			val := v.ValueString()
			return platform.StateRoleEnum(val)
		}),
	}
	return result
}

func (s State) updateActions(plan State) platform.StateUpdate {
	result := platform.StateUpdate{
		Version: int(s.Version.ValueInt64()),
		Actions: []platform.StateUpdateAction{},
	}

	// setName
	if !reflect.DeepEqual(s.Name, plan.Name) {
		result.Actions = append(
			result.Actions,
			platform.StateSetNameAction{
				Name: plan.Name.ValueLocalizedString(),
			})
	}

	// setDescription
	if !reflect.DeepEqual(s.Description, plan.Description) {
		spew.Dump(s.Description, plan.Description)
		result.Actions = append(
			result.Actions,
			platform.StateSetDescriptionAction{
				Description: plan.Description.ValueLocalizedString(),
			})
	}

	// changeKey
	if s.Key != plan.Key {
		result.Actions = append(
			result.Actions,
			platform.StateChangeKeyAction{Key: plan.Key.ValueString()})
	}

	// changeType
	if !s.Type.Equal(plan.Type) {
		result.Actions = append(
			result.Actions,
			platform.StateChangeTypeAction{Type: platform.StateTypeEnum(plan.Type.ValueString())})
	}

	// changeInitial
	if !s.Initial.Equal(plan.Initial) {
		result.Actions = append(
			result.Actions,
			platform.StateChangeInitialAction{
				Initial: plan.Initial.ValueBool(),
			})
	}

	// setRoles
	if !reflect.DeepEqual(s.Roles, plan.Roles) {
		roles := pie.Map(plan.Roles, func(v types.String) platform.StateRoleEnum {
			val := v.ValueString()
			return platform.StateRoleEnum(val)
		})
		result.Actions = append(
			result.Actions,
			platform.StateSetRolesAction{Roles: roles})
	}

	return result
}

func (s *State) matchDefaults(state State) {
	// If the remote state value for initial is false and the plan/state
	// has nil, then set it to false (since it's the default)
	if !s.Initial.ValueBool() && state.Initial.IsNull() {
		s.Initial = state.Initial
	}
}

func (s *State) setDefaults() {
	if s.Initial.IsNull() {
		s.Initial = types.BoolValue(false)
	}
}
