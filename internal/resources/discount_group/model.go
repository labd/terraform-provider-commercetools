package discount_group

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// DiscountGroup is the main resource schema data
type DiscountGroup struct {
	ID          types.String                     `tfsdk:"id"`
	Version     types.Int64                      `tfsdk:"version"`
	Key         types.String                     `tfsdk:"key"`
	Name        customtypes.LocalizedStringValue `tfsdk:"name"`
	Description customtypes.LocalizedStringValue `tfsdk:"description"`
	SortOrder   types.String                     `tfsdk:"sort_order"`
	IsActive    types.Bool                       `tfsdk:"is_active"`
}

func toDraft(g *DiscountGroup) platform.DiscountGroupDraft {
	return platform.DiscountGroupDraft{
		Key:         g.Key.ValueString(),
		Name:        g.Name.ValueLocalizedStringRef(),
		Description: g.Description.ValueLocalizedStringRef(),
		SortOrder:   g.SortOrder.ValueString(),
		IsActive:    g.IsActive.ValueBoolPointer(),
	}
}

func fromNative(n *platform.DiscountGroup) DiscountGroup {
	return DiscountGroup{
		ID:          types.StringValue(n.ID),
		Version:     types.Int64Value(int64(n.Version)),
		Key:         types.StringValue(n.Key),
		Name:        utils.FromOptionalLocalizedString(n.Name),
		Description: utils.FromOptionalLocalizedString(n.Description),
		SortOrder:   types.StringValue(n.SortOrder),
		IsActive:    types.BoolValue(n.IsActive),
	}
}

func toUpdateActions(state *DiscountGroup, plan *DiscountGroup) platform.DiscountGroupUpdate {
	update := platform.DiscountGroupUpdate{
		Version: int(state.Version.ValueInt64()),
		Actions: []platform.DiscountGroupUpdateAction{},
	}

	if !state.Key.Equal(plan.Key) {
		update.Actions = append(update.Actions, platform.DiscountGroupSetKeyAction{
			Key: plan.Key.ValueString(),
		})
	}

	if !reflect.DeepEqual(state.Name, plan.Name) {
		update.Actions = append(update.Actions, platform.DiscountGroupSetNameAction{
			Name: plan.Name.ValueLocalizedStringRef(),
		})
	}

	if !reflect.DeepEqual(state.Description, plan.Description) {
		update.Actions = append(update.Actions, platform.DiscountGroupSetDescriptionAction{
			Description: plan.Description.ValueLocalizedStringRef(),
		})
	}

	if !state.SortOrder.Equal(plan.SortOrder) {
		update.Actions = append(update.Actions, platform.DiscountGroupSetSortOrderAction{
			SortOrder: plan.SortOrder.ValueString(),
		})
	}

	if !state.IsActive.Equal(plan.IsActive) {
		update.Actions = append(update.Actions, platform.DiscountGroupSetIsActiveAction{
			IsActive: plan.IsActive.ValueBool(),
		})
	}

	return update
}
