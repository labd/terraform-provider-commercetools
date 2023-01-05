package project

import (
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
						Enabled: types.BoolValue(false),
					},
				},
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
						Label: map[string]types.String{
							"nl": types.StringValue("licht"),
							"en": types.StringValue("light"),
						},
					},
				},
			},
			action: platform.ProjectUpdate{
				Version: 1,
				Actions: []platform.ProjectUpdateAction{
					platform.ProjectChangeNameAction{
						Name: "my new name",
					},
					platform.ProjectChangeCountriesAction{
						Countries: []string{"NL", "DE"},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.UpdateActions(tt.plan)
			assert.Equal(t, tt.action, result)
		})
	}
}
