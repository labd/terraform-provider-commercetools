package product_selection

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestProductSelection_UpdateActions(t *testing.T) {
	cases := []struct {
		name     string
		state    ProductSelection
		plan     ProductSelection
		expected platform.ProductSelectionUpdate
	}{
		{
			"product selection update name",
			ProductSelection{
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Example product selection"),
				}),
			},
			ProductSelection{
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("Example other product selection"),
				}),
			},
			platform.ProductSelectionUpdate{
				Actions: []platform.ProductSelectionUpdateAction{
					platform.ProductSelectionChangeNameAction{
						Name: map[string]string{"en-US": "Example other product selection"},
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

func TestNewProductSelectionFromNative(t *testing.T) {
	cases := []struct {
		name   string
		res    *platform.ProductSelection
		expect ProductSelection
	}{
		{
			"decode remote product selection representation into local resource",
			&platform.ProductSelection{
				ID:      "rand-uuid-or-other-string",
				Version: 1,
				Key:     utils.StringRef("ps-1"),
				Name:    map[string]string{"en-US": "the selection"},
				Mode:    platform.ProductSelectionModeIndividual,
			},
			ProductSelection{
				ID:      types.StringValue("rand-uuid-or-other-string"),
				Key:     types.StringValue("ps-1"),
				Version: types.Int64Value(1),
				Name: customtypes.NewLocalizedStringValue(map[string]attr.Value{
					"en-US": types.StringValue("the selection"),
				}),
				Mode: types.StringValue("Individual"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := NewProductSelectionFromNative(c.res)
			assert.EqualValues(t, got, c.expect)
		})
	}
}
