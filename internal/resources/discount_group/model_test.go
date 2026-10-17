package discount_group

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
)

func TestToUpdateActions(t *testing.T) {
	state := DiscountGroup{
		Version:   types.Int64Value(2),
		Key:       types.StringValue("summer"),
		Name:      customtypes.NewLocalizedStringValue(map[string]attr.Value{"en": types.StringValue("Summer")}),
		SortOrder: types.StringValue("0.6"),
		IsActive:  types.BoolValue(true),
	}

	t.Run("nothing changed", func(t *testing.T) {
		assert.Empty(t, toUpdateActions(&state, &state).Actions)
	})

	t.Run("every field changed", func(t *testing.T) {
		plan := state
		plan.Key = types.StringValue("winter")
		plan.Name = customtypes.NewLocalizedStringValue(map[string]attr.Value{"en": types.StringValue("Winter")})
		plan.SortOrder = types.StringValue("0.7")
		plan.IsActive = types.BoolValue(false)

		update := toUpdateActions(&state, &plan)
		assert.Equal(t, 2, update.Version)
		assert.Equal(t, []platform.DiscountGroupUpdateAction{
			platform.DiscountGroupSetKeyAction{Key: "winter"},
			platform.DiscountGroupSetNameAction{Name: &platform.LocalizedString{"en": "Winter"}},
			platform.DiscountGroupSetSortOrderAction{SortOrder: "0.7"},
			platform.DiscountGroupSetIsActiveAction{IsActive: false},
		}, update.Actions)
	})
}
