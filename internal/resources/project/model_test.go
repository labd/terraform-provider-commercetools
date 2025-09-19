package project

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"

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

				EnableSearchIndexProducts:      types.BoolValue(false),
				EnableSearchIndexOrders:        types.BoolValue(false),
				EnableSearchIndexCustomers:     types.BoolValue(false),
				EnableSearchIndexProductSearch: types.BoolValue(false),
				EnableSearchIndexBusinessUnits: types.BoolValue(false),

				ExternalOAuth: []ExternalOAuth{},
				Carts:         nil,
				Messages: []Messages{
					{
						Enabled:                 types.BoolValue(false),
						DeleteDaysAfterCreation: types.Int64Value(DefaultDeleteDaysAfterCreation),
					},
				},
				ShippingRateCartClassificationValue: []models.CustomFieldLocalizedEnumValue{},
				BusinessUnits:                       []BusinessUnits{},
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
						PriceRoundingMode:               types.StringValue("HalfUp"),
						TaxRoundingMode:                 types.StringValue("HalfUp"),
					},
				},
			},
			plan: Project{
				Version: types.Int64Value(1),
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolValue(false),
						DeleteDaysAfterLastModification: types.Int64Value(90),
						PriceRoundingMode:               types.StringValue("HalfEven"),
						TaxRoundingMode:                 types.StringValue("HalfEven"),
					},
				},
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeCartsConfigurationAction{
						CartsConfiguration: platform.CartsConfiguration{
							DeleteDaysAfterLastModification: utils.Ref(90),
							CountryTaxRateFallbackEnabled:   utils.Ref(false),
							PriceRoundingMode:               utils.Ref(platform.RoundingModeHalfEven),
							TaxRoundingMode:                 utils.Ref(platform.RoundingModeHalfEven),
						},
					},
					platform.ProjectChangeCountryTaxRateFallbackEnabledAction{CountryTaxRateFallbackEnabled: false},
					platform.ProjectChangePriceRoundingModeAction{PriceRoundingMode: "HalfEven"},
					platform.ProjectChangeTaxRoundingModeAction{TaxRoundingMode: "HalfEven"},
				},
			},
		},
		{
			name: "Update with search index orders activated",
			state: Project{
				Version:                 types.Int64Value(1),
				EnableSearchIndexOrders: types.BoolValue(false),
			},
			plan: Project{
				Version: types.Int64Value(1),

				EnableSearchIndexOrders: types.BoolValue(true),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeOrderSearchStatusAction{Status: platform.OrderSearchStatusActivated},
				},
			},
		},
		{
			name: "Update with search index orders deactivated",
			state: Project{
				Version:                 types.Int64Value(1),
				EnableSearchIndexOrders: types.BoolValue(true),
			},
			plan: Project{
				Version:                 types.Int64Value(1),
				EnableSearchIndexOrders: types.BoolValue(false),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeOrderSearchStatusAction{Status: platform.OrderSearchStatusDeactivated},
				},
			},
		},
		{
			name: "Update with search index orders no changes",
			state: Project{
				Version:                 types.Int64Value(1),
				EnableSearchIndexOrders: types.BoolValue(false),
			},
			plan: Project{
				Version:                 types.Int64Value(1),
				EnableSearchIndexOrders: types.BoolValue(false),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{},
			},
		},
		{
			name: "Update with search index customers activated",
			state: Project{
				Version:                    types.Int64Value(1),
				EnableSearchIndexCustomers: types.BoolValue(false),
			},
			plan: Project{
				Version:                    types.Int64Value(1),
				EnableSearchIndexCustomers: types.BoolValue(true),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeCustomerSearchStatusAction{Status: platform.CustomerSearchStatusActivated},
				},
			},
		},
		{
			name: "Update with search index customers deactivated",
			state: Project{
				Version:                    types.Int64Value(1),
				EnableSearchIndexCustomers: types.BoolValue(true),
			},
			plan: Project{
				Version:                    types.Int64Value(1),
				EnableSearchIndexCustomers: types.BoolValue(false),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeCustomerSearchStatusAction{Status: platform.CustomerSearchStatusDeactivated},
				},
			},
		},
		{
			name: "Update with search index customers no changes",
			state: Project{
				Version:                    types.Int64Value(1),
				EnableSearchIndexCustomers: types.BoolValue(false),
			},
			plan: Project{
				Version:                    types.Int64Value(1),
				EnableSearchIndexCustomers: types.BoolValue(false),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{},
			},
		},
		{
			name: "Update with search index products activated",
			state: Project{
				Version:                   types.Int64Value(1),
				EnableSearchIndexProducts: types.BoolValue(false),
			},
			plan: Project{
				Version:                   types.Int64Value(1),
				EnableSearchIndexProducts: types.BoolValue(true),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeProductSearchIndexingEnabledAction{
						Enabled: true,
						Mode:    utils.GetRef(platform.ProductSearchIndexingModeProductProjectionsSearch),
					},
				},
			},
		},
		{
			name: "Update with search index products deactivated",
			state: Project{
				Version:                   types.Int64Value(1),
				EnableSearchIndexProducts: types.BoolValue(true),
			},
			plan: Project{
				Version:                   types.Int64Value(1),
				EnableSearchIndexProducts: types.BoolValue(false),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeProductSearchIndexingEnabledAction{
						Enabled: false,
						Mode:    utils.GetRef(platform.ProductSearchIndexingModeProductProjectionsSearch),
					},
				},
			},
		},
		{
			name: "Update with search index product search activated",
			state: Project{
				Version:                        types.Int64Value(1),
				EnableSearchIndexProductSearch: types.BoolValue(false),
			},
			plan: Project{
				Version:                        types.Int64Value(1),
				EnableSearchIndexProductSearch: types.BoolValue(true),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeProductSearchIndexingEnabledAction{
						Enabled: true,
						Mode:    utils.GetRef(platform.ProductSearchIndexingModeProductsSearch),
					},
				},
			},
		},
		{
			name: "Update with search index product search deactivated",
			state: Project{
				Version:                        types.Int64Value(1),
				EnableSearchIndexProductSearch: types.BoolValue(true),
			},
			plan: Project{
				Version:                        types.Int64Value(1),
				EnableSearchIndexProductSearch: types.BoolValue(false),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeProductSearchIndexingEnabledAction{
						Enabled: false,
						Mode:    utils.GetRef(platform.ProductSearchIndexingModeProductsSearch),
					},
				},
			},
		},
		{
			name: "Create with business unit settings",
			state: Project{
				Version:       types.Int64Value(1),
				BusinessUnits: []BusinessUnits{},
			},
			plan: Project{
				Version: types.Int64Value(1),
				BusinessUnits: []BusinessUnits{
					{
						MyBusinessUnitStatusOnCreation:           types.StringValue(string(platform.BusinessUnitConfigurationStatusActive)),
						MyBusinessUnitAssociateRoleKeyOnCreation: types.StringValue("my-associate-role"),
					},
				},
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeBusinessUnitStatusOnCreationAction{
						Status: platform.BusinessUnitConfigurationStatusActive,
					},
					platform.ProjectSetBusinessUnitAssociateRoleOnCreationAction{
						AssociateRole: platform.AssociateRoleResourceIdentifier{
							Key: utils.StringRef("my-associate-role"),
						},
					},
				},
			},
		},
		{
			name: "Update with search index business unit search activated",
			state: Project{
				Version:                        types.Int64Value(1),
				EnableSearchIndexBusinessUnits: types.BoolValue(false),
			},
			plan: Project{
				Version:                        types.Int64Value(1),
				EnableSearchIndexBusinessUnits: types.BoolValue(true),
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeBusinessUnitSearchStatusAction{
						Status: platform.BusinessUnitSearchStatusActivated,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.state.updateActions(tt.plan)
			assert.NoError(t, err)
			assert.Equal(t, tt.action, result)
		})
	}
}

func TestSetStateData(t *testing.T) {
	tests := []struct {
		name     string
		state    Project
		plan     Project
		expected Project
	}{
		{
			name: "externalOAuth nil",
			state: Project{
				ExternalOAuth: nil,
				Carts: []Carts{
					{},
				},
			},
			plan: Project{
				ExternalOAuth: nil,
			},
			expected: Project{
				ExternalOAuth: nil,
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolNull(),
						DeleteDaysAfterLastModification: types.Int64Null(),
						PriceRoundingMode:               types.StringNull(),
						TaxRoundingMode:                 types.StringNull(),
					},
				},
			},
		}, {
			name: "externalOAuth in state",
			state: Project{
				ExternalOAuth: []ExternalOAuth{
					{AuthorizationHeader: types.StringValue("some-value")},
				},
				Carts: []Carts{
					{},
				},
			},
			plan: Project{
				ExternalOAuth: nil,
			},
			expected: Project{
				ExternalOAuth: []ExternalOAuth{
					{AuthorizationHeader: types.StringValue("some-value")},
				},
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolNull(),
						DeleteDaysAfterLastModification: types.Int64Null(),
						PriceRoundingMode:               types.StringNull(),
						TaxRoundingMode:                 types.StringNull(),
					},
				},
			},
		}, {
			name: "externalOAuth in plan",
			state: Project{
				ExternalOAuth: []ExternalOAuth{
					{AuthorizationHeader: types.StringValue("some-value")},
				},
				Carts: []Carts{
					{},
				},
			},
			plan: Project{
				ExternalOAuth: []ExternalOAuth{
					{AuthorizationHeader: types.StringValue("some-other-value")},
				},
			},
			expected: Project{
				ExternalOAuth: []ExternalOAuth{
					{AuthorizationHeader: types.StringValue("some-other-value")},
				},
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolNull(),
						DeleteDaysAfterLastModification: types.Int64Null(),
						PriceRoundingMode:               types.StringNull(),
						TaxRoundingMode:                 types.StringNull(),
					},
				},
			},
		}, {
			name: "business unit in plan",
			state: Project{
				BusinessUnits: []BusinessUnits{
					{
						MyBusinessUnitStatusOnCreation: types.StringValue(string(platform.BusinessUnitConfigurationStatusInactive)),
					},
				},
				Carts: []Carts{
					{},
				},
			},
			plan: Project{
				BusinessUnits: []BusinessUnits{},
			},
			expected: Project{
				BusinessUnits: nil,
				Carts: []Carts{
					{
						CountryTaxRateFallbackEnabled:   types.BoolNull(),
						DeleteDaysAfterLastModification: types.Int64Null(),
						PriceRoundingMode:               types.StringNull(),
						TaxRoundingMode:                 types.StringNull(),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.setStateData(tt.plan)
			assert.Equal(t, tt.expected, tt.state)
		})
	}
}

func IsDefaultCartsConfiguration_DefaultValues(t *testing.T) {
	c := platform.CartsConfiguration{
		CountryTaxRateFallbackEnabled:   utils.BoolRef(false),
		DeleteDaysAfterLastModification: utils.IntRef(DefaultDaysAfterLastModification),
		PriceRoundingMode:               utils.GetRef(platform.RoundingModeHalfEven),
		TaxRoundingMode:                 utils.GetRef(platform.RoundingModeHalfEven),
	}
	assert.True(t, IsDefaultCartsConfiguration(c))
}

func IsDefaultCartsConfiguration_NilFields(t *testing.T) {
	c := platform.CartsConfiguration{}
	assert.True(t, IsDefaultCartsConfiguration(c))
}

func IsDefaultCartsConfiguration_NonDefaultCountryTaxRateFallbackEnabled(t *testing.T) {
	c := platform.CartsConfiguration{
		CountryTaxRateFallbackEnabled: utils.BoolRef(true),
	}
	assert.False(t, IsDefaultCartsConfiguration(c))
}

func IsDefaultCartsConfiguration_NonDefaultDeleteDaysAfterLastModification(t *testing.T) {
	c := platform.CartsConfiguration{
		DeleteDaysAfterLastModification: utils.IntRef(30),
	}
	assert.False(t, IsDefaultCartsConfiguration(c))
}

func IsDefaultCartsConfiguration_NonDefaultPriceRoundingMode(t *testing.T) {
	c := platform.CartsConfiguration{
		PriceRoundingMode: utils.GetRef(platform.RoundingModeHalfUp),
	}
	assert.False(t, IsDefaultCartsConfiguration(c))
}

func IsDefaultCartsConfiguration_NonDefaultTaxRoundingMode(t *testing.T) {
	c := platform.CartsConfiguration{
		TaxRoundingMode: utils.GetRef(platform.RoundingModeHalfUp),
	}
	assert.False(t, IsDefaultCartsConfiguration(c))
}

func IsDefaultShoppingListsConfiguration_Nil(t *testing.T) {
	assert.True(t, IsDefaultShoppingListsConfiguration(nil))
}

func IsDefaultShoppingListsConfiguration_DefaultValue(t *testing.T) {
	c := &platform.ShoppingListsConfiguration{
		DeleteDaysAfterLastModification: utils.IntRef(DefaultDaysAfterLastModification),
	}
	assert.True(t, IsDefaultShoppingListsConfiguration(c))
}

func IsDefaultShoppingListsConfiguration_NilDeleteDays(t *testing.T) {
	c := &platform.ShoppingListsConfiguration{}
	assert.True(t, IsDefaultShoppingListsConfiguration(c))
}

func IsDefaultShoppingListsConfiguration_NonDefaultDeleteDays(t *testing.T) {
	c := &platform.ShoppingListsConfiguration{
		DeleteDaysAfterLastModification: utils.IntRef(30),
	}
	assert.False(t, IsDefaultShoppingListsConfiguration(c))
}
