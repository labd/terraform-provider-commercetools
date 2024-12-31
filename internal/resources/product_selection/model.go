package product_selection

import (
	"github.com/labd/terraform-provider-commercetools/internal/sharedtypes"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// ProductSelection represents the main schema data.
type ProductSelection struct {
	ID      types.String                     `tfsdk:"id"`
	Key     types.String                     `tfsdk:"key"`
	Version types.Int64                      `tfsdk:"version"`
	Name    customtypes.LocalizedStringValue `tfsdk:"name"`
	Mode    types.String                     `tfsdk:"mode"`
	Custom  *sharedtypes.Custom              `tfsdk:"custom"`
}

func NewProductSelectionFromNative(ps *platform.ProductSelection) (ProductSelection, error) {
	custom, err := sharedtypes.NewCustomFromNative(ps.Custom)
	if err != nil {
		return ProductSelection{}, err
	}

	return ProductSelection{
		ID:      types.StringValue(ps.ID),
		Version: types.Int64Value(int64(ps.Version)),
		Name:    utils.FromLocalizedString(ps.Name),
		Key:     utils.FromOptionalString(ps.Key),
		Mode:    types.StringValue(string(ps.Mode)),
		Custom:  custom,
	}, err
}

func (ps ProductSelection) draft(t *platform.Type) (platform.ProductSelectionDraft, error) {
	custom, err := ps.Custom.Draft(t)
	if err != nil {
		return platform.ProductSelectionDraft{}, err
	}

	return platform.ProductSelectionDraft{
		Key:    ps.Key.ValueStringPointer(),
		Name:   ps.Name.ValueLocalizedString(),
		Mode:   (*platform.ProductSelectionMode)(ps.Mode.ValueStringPointer()),
		Custom: custom,
	}, nil
}

func (ps ProductSelection) updateActions(t *platform.Type, plan ProductSelection) (platform.ProductSelectionUpdate, error) {
	result := platform.ProductSelectionUpdate{
		Version: int(ps.Version.ValueInt64()),
		Actions: []platform.ProductSelectionUpdateAction{},
	}

	// setName
	if !reflect.DeepEqual(ps.Name, plan.Name) {
		result.Actions = append(
			result.Actions,
			platform.ProductSelectionChangeNameAction{
				Name: plan.Name.ValueLocalizedString(),
			})
	}

	// changeKey
	if ps.Key != plan.Key {
		result.Actions = append(
			result.Actions,
			platform.ProductSelectionSetKeyAction{Key: plan.Key.ValueStringPointer()})
	}

	// setCustomFields
	if !reflect.DeepEqual(ps.Custom, plan.Custom) {
		actions, err := sharedtypes.CustomFieldUpdateActions[
			platform.ProductSelectionSetCustomTypeAction,
			platform.ProductSelectionSetCustomFieldAction,
		](t, ps.Custom, plan.Custom)
		if err != nil {
			return platform.ProductSelectionUpdate{}, err
		}

		for i := range actions {
			result.Actions = append(result.Actions, actions[i].(platform.AssociateRoleUpdateAction))
		}
	}

	return result, nil
}
