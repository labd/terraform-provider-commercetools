package associate_role

import (
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// AssociateRole represents the main schema data.
type AssociateRole struct {
	ID              types.String        `tfsdk:"id"`
	Key             types.String        `tfsdk:"key"`
	Version         types.Int64         `tfsdk:"version"`
	Name            types.String        `tfsdk:"name"`
	BuyerAssignable types.Bool          `tfsdk:"buyer_assignable"`
	Permissions     []types.String      `tfsdk:"permissions"`
	Custom          *sharedtypes.Custom `tfsdk:"custom"`
}

func NewAssociateRoleFromNative(ar *platform.AssociateRole) (AssociateRole, error) {
	custom, err := sharedtypes.NewCustomFromNative(ar.Custom)
	if err != nil {
		return AssociateRole{}, err
	}

	return AssociateRole{
		ID:              types.StringValue(ar.ID),
		Version:         types.Int64Value(int64(ar.Version)),
		Name:            types.StringValue(*ar.Name),
		Key:             types.StringValue(ar.Key),
		BuyerAssignable: types.BoolValue(ar.BuyerAssignable),
		Permissions: pie.Map(ar.Permissions, func(perm platform.Permission) types.String {
			return types.StringValue(string(perm))
		}),
		Custom: custom,
	}, nil
}

func (ar *AssociateRole) draft(t *platform.Type) (platform.AssociateRoleDraft, error) {
	custom, err := ar.Custom.Draft(t)
	if err != nil {
		return platform.AssociateRoleDraft{}, err
	}

	return platform.AssociateRoleDraft{
		Key:             ar.Key.ValueString(),
		Name:            ar.Name.ValueStringPointer(),
		BuyerAssignable: ar.BuyerAssignable.ValueBoolPointer(),
		Permissions: pie.Map(ar.Permissions, func(p types.String) platform.Permission {
			return platform.Permission(p.ValueString())
		}),
		Custom: custom,
	}, nil
}

func (ar *AssociateRole) updateActions(t *platform.Type, plan AssociateRole) (platform.AssociateRoleUpdate, error) {
	result := platform.AssociateRoleUpdate{
		Version: int(ar.Version.ValueInt64()),
		Actions: []platform.AssociateRoleUpdateAction{},
	}

	// setName
	if !ar.Name.Equal(plan.Name) {
		var newName *string
		if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
			newName = utils.StringRef(plan.Name.ValueString())
		}

		result.Actions = append(
			result.Actions,
			platform.AssociateRoleSetNameAction{Name: newName},
		)
	}

	// setBuyerAssignable value
	if !ar.BuyerAssignable.Equal(plan.BuyerAssignable) {
		result.Actions = append(
			result.Actions,
			platform.AssociateRoleChangeBuyerAssignableAction{
				BuyerAssignable: plan.BuyerAssignable.ValueBool(),
			},
		)
	}

	// setNewOrRemovedPermissions
	if !reflect.DeepEqual(ar.Permissions, plan.Permissions) {
		// we completely override the values as calculating
		// differences will overcomplicate operations.
		result.Actions = append(
			result.Actions,
			platform.AssociateRoleSetPermissionsAction{
				Permissions: pie.Map(plan.Permissions, func(p types.String) platform.Permission {
					return platform.Permission(p.ValueString())
				}),
			},
		)
	}

	// setCustomFields
	if !reflect.DeepEqual(ar.Custom, plan.Custom) {
		actions, err := sharedtypes.CustomFieldUpdateActions[
			platform.AssociateRoleSetCustomTypeAction,
			platform.AssociateRoleSetCustomFieldAction,
		](t, ar.Custom, plan.Custom)
		if err != nil {
			return platform.AssociateRoleUpdate{}, err
		}

		for i := range actions {
			result.Actions = append(result.Actions, actions[i].(platform.AssociateRoleUpdateAction))
		}
	}

	return result, nil
}
