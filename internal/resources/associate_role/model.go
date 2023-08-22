package associate_role

import (
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// AssociateRole represents the main schema data.
type AssociateRole struct {
	ID              types.String   `tfsdk:"id"`
	Key             types.String   `tfsdk:"key"`
	Version         types.Int64    `tfsdk:"version"`
	Name            types.String   `tfsdk:"name"`
	BuyerAssignable types.Bool     `tfsdk:"buyer_assignable"`
	Permissions     []types.String `tfsdk:"permissions"`
}

func NewAssociateRoleFromNative(ar *platform.AssociateRole) AssociateRole {
	return AssociateRole{
		ID:              types.StringValue(ar.ID),
		Version:         types.Int64Value(int64(ar.Version)),
		Name:            types.StringValue(*ar.Name),
		Key:             types.StringValue(ar.Key),
		BuyerAssignable: types.BoolValue(ar.BuyerAssignable),
		Permissions: pie.Map(ar.Permissions, func(perm platform.Permission) types.String {
			return types.StringValue(string(perm))
		}),
	}
}

func (ar AssociateRole) draft() platform.AssociateRoleDraft {
	return platform.AssociateRoleDraft{
		Key:             ar.Key.ValueString(),
		Name:            ar.Name.ValueStringPointer(),
		BuyerAssignable: ar.BuyerAssignable.ValueBoolPointer(),
		Permissions: pie.Map(ar.Permissions, func(p types.String) platform.Permission {
			return platform.Permission(p.ValueString())
		}),
	}
}

func (ar AssociateRole) updateActions(plan AssociateRole) platform.AssociateRoleUpdate {
	result := platform.AssociateRoleUpdate{
		Version: int(ar.Version.ValueInt64()),
		Actions: []platform.AssociateRoleUpdateAction{},
	}

	// setName
	if ar.Name != plan.Name {
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
	if ar.BuyerAssignable != plan.BuyerAssignable {
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

	return result
}
