package product_selection

import (
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
}

func NewProductSelectionFromNative(ps *platform.ProductSelection) ProductSelection {
	return ProductSelection{
		ID:      types.StringValue(ps.ID),
		Version: types.Int64Value(int64(ps.Version)),
		Name:    utils.FromLocalizedString(ps.Name),
		Key:     utils.FromOptionalString(ps.Key),
		Mode:    types.StringValue(string(ps.Mode)),
	}
}

func (ps ProductSelection) draft() platform.ProductSelectionDraft {
	return platform.ProductSelectionDraft{
		Key:  ps.Key.ValueStringPointer(),
		Name: ps.Name.ValueLocalizedString(),
		Mode: (*platform.ProductSelectionMode)(ps.Mode.ValueStringPointer()),
	}
}

func (ps ProductSelection) updateActions(plan ProductSelection) platform.ProductSelectionUpdate {
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

	return result
}
