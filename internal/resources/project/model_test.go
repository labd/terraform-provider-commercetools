package project

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/models"
)

func TestNewProjectFromNative(t *testing.T) {
	tests := []struct {
		name string
		res  *platform.Project
		want Project
	}{
		{
			name: "Default",
			res: &platform.Project{
				Version: 1,
				Key:     "my-project",
				Name:    "my project",
			},
			want: Project{
				Version: types.Int64Value(1),
				ID:      types.StringValue("my-project"),
				Key:     types.StringValue("my-project"),
				Name:    types.StringValue("my project"),

				EnableSearchIndexProducts: types.BoolValue(false),
				EnableSearchIndexOrders:   types.BoolValue(false),

				ExternalOAuth: []ExternalOAuth{},
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolNull(),
						DeleteDaysAfterLastModification: types.Int64Null(),
					},
				},
				Messages: []Messages{
					{
						Enabled:                 types.BoolValue(false),
						DeleteDaysAfterCreation: types.Int64Value(DefaultDeleteDaysAfterCreation),
					},
				},
				ShippingRateCartClassificationValue: []models.CustomFieldLocalizedEnumValue{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewProjectFromNative(tt.res)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestUpdateActions(t *testing.T) {
	tests := []struct {
		name   string
		state  Project
		plan   Project
		action platform.ProjectUpdate
	}{
		{
			name: "Default",
			state: Project{
				Version:               types.Int64Value(1),
				ID:                    types.StringValue("my-project"),
				Key:                   types.StringValue("my-project"),
				Name:                  types.StringValue("my project"),
				Countries:             []types.String{types.StringValue("US")},
				ShippingRateInputType: types.StringValue("CartValue"),
			},
			plan: Project{
				Version: types.Int64Value(1),
				ID:      types.StringValue("my-project"),
				Key:     types.StringValue("my-project"),
				Name:    types.StringValue("my new name"),
				Countries: []types.String{
					types.StringValue("NL"),
					types.StringValue("DE"),
				},
				ShippingRateInputType: types.StringValue("CartClassification"),
				ShippingRateCartClassificationValue: []models.CustomFieldLocalizedEnumValue{
					{
						Key: types.StringValue("Light"),
						Label: customtypes.NewLocalizedStringValue(map[string]attr.Value{
							"nl": types.StringValue("licht"),
							"en": types.StringValue("light"),
						}),
					},
				},
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeCountriesAction{
						Countries: []string{"NL", "DE"},
					},
					platform.ProjectChangeNameAction{
						Name: "my new name",
					},
					platform.ProjectSetShippingRateInputTypeAction{
						ShippingRateInputType: platform.CartClassificationType{
							Values: []platform.CustomFieldLocalizedEnumValue{
								{
									Key: "Light",
									Label: platform.LocalizedString{
										"en": "light",
										"nl": "licht",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Carts Configuration",
			state: Project{
				Version: types.Int64Value(1),
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolValue(true),
						DeleteDaysAfterLastModification: types.Int64Value(10),
					},
				},
			},
			plan: Project{
				Version: types.Int64Value(1),
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolValue(false),
						DeleteDaysAfterLastModification: types.Int64Value(90),
					},
				},
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeCartsConfigurationAction{
						CartsConfiguration: platform.CartsConfiguration{
							CountryTaxRateFallbackEnabled:   utils.BoolRef(false),
							DeleteDaysAfterLastModification: utils.IntRef(90),
						},
					},
					platform.ProjectChangeCountryTaxRateFallbackEnabledAction{CountryTaxRateFallbackEnabled: false},
				},
			},
		},
		{
			name: "Create with bool unknown",
			state: Project{
				Version:                   types.Int64Value(1),
				EnableSearchIndexOrders:   types.BoolValue(false),
				EnableSearchIndexProducts: types.BoolValue(false),
			},
			plan: Project{
				Version: types.Int64Value(1),

				EnableSearchIndexOrders:   types.BoolValue(true),
				EnableSearchIndexProducts: types.BoolUnknown(),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeOrderSearchStatusAction{Status: platform.OrderSearchStatusActivated},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.updateActions(tt.plan)
			assert.Equal(t, tt.action, result)
		})
	}
}
