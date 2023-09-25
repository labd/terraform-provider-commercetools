package project

import (
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/models"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

const (
	DefaultDeleteDaysAfterCreation = 15
)

type Project struct {
	ID      types.String `tfsdk:"id"`
	Key     types.String `tfsdk:"key"`
	Version types.Int64  `tfsdk:"version"`

	Name       types.String   `tfsdk:"name"`
	Currencies []types.String `tfsdk:"currencies"`
	Countries  []types.String `tfsdk:"countries"`
	Languages  []types.String `tfsdk:"languages"`

	EnableSearchIndexProducts types.Bool `tfsdk:"enable_search_index_products"`
	EnableSearchIndexOrders   types.Bool `tfsdk:"enable_search_index_orders"`

	// These items all have maximal one item. We don't use SingleNestedBlock
	// here since it isn't quite robust currently.
	// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
	Carts         []Carts         `tfsdk:"carts"`
	Messages      []Messages      `tfsdk:"messages"`
	ExternalOAuth []ExternalOAuth `tfsdk:"external_oauth"`

	ShippingRateInputType               types.String                           `tfsdk:"shipping_rate_input_type"`
	ShippingRateCartClassificationValue []models.CustomFieldLocalizedEnumValue `tfsdk:"shipping_rate_cart_classification_value"`
}

func NewProjectFromNative(n *platform.Project) Project {
	res := Project{
		Version: types.Int64Value(int64(n.Version)),
		ID:      types.StringValue(n.Key),
		Key:     types.StringValue(n.Key),
		Name:    types.StringValue(n.Name),

		Currencies: pie.Map(n.Currencies, types.StringValue),
		Countries:  pie.Map(n.Countries, types.StringValue),
		Languages:  pie.Map(n.Languages, types.StringValue),

		EnableSearchIndexProducts: types.BoolValue(false),
		EnableSearchIndexOrders:   types.BoolValue(false),

		Carts: []Carts{
			{
				DeleteDaysAfterLastModification: utils.FromOptionalInt(n.Carts.DeleteDaysAfterLastModification),
				CountryTaxRateFallbackEnabled:   utils.FromOptionalBool(n.Carts.CountryTaxRateFallbackEnabled),
			},
		},
		Messages: []Messages{
			{
				DeleteDaysAfterCreation: utils.FromOptionalInt(n.Messages.DeleteDaysAfterCreation),
				Enabled:                 types.BoolValue(n.Messages.Enabled),
			},
		},
		ExternalOAuth: []ExternalOAuth{},
	}

	// always set it to an empty list to avoid the wrong comparison in the update actions part
	res.ShippingRateCartClassificationValue = []models.CustomFieldLocalizedEnumValue{}

	switch s := n.ShippingRateInputType.(type) {
	case platform.CartScoreType:
		res.ShippingRateInputType = types.StringValue("CartScore")
	case platform.CartValueType:
		res.ShippingRateInputType = types.StringValue("CartValue")
	case platform.CartClassificationType:
		res.ShippingRateInputType = types.StringValue("CartClassification")
		values := make([]models.CustomFieldLocalizedEnumValue, len(s.Values))
		for i := range s.Values {
			values[i] = models.NewCustomFieldLocalizedEnumValue(s.Values[i])
		}
		res.ShippingRateCartClassificationValue = values
	}

	// If delete_days_after_creation is nil (before version 1.6) then we set it
	// to the commercetools default of 15
	if res.Messages[0].DeleteDaysAfterCreation.IsNull() {
		res.Messages[0].DeleteDaysAfterCreation = types.Int64Value(DefaultDeleteDaysAfterCreation)
	}

	if n.SearchIndexing != nil && n.SearchIndexing.Products != nil && n.SearchIndexing.Products.Status != nil {
		status := *n.SearchIndexing.Products.Status
		enabled := status != platform.SearchIndexingConfigurationStatusDeactivated
		res.EnableSearchIndexProducts = types.BoolValue(enabled)
	}

	if n.SearchIndexing != nil && n.SearchIndexing.Orders != nil && n.SearchIndexing.Orders.Status != nil {
		status := *n.SearchIndexing.Orders.Status
		enabled := status != platform.SearchIndexingConfigurationStatusDeactivated
		res.EnableSearchIndexOrders = types.BoolValue(enabled)
	}

	if n.ExternalOAuth != nil {
		res.ExternalOAuth = []ExternalOAuth{
			{
				URL:                 types.StringValue(n.ExternalOAuth.Url),
				AuthorizationHeader: types.StringUnknown(),
			},
		}
	}

	return res
}

func (p *Project) setStateData(o Project) {
	if len(p.ExternalOAuth) > 0 {
		p.ExternalOAuth[0].AuthorizationHeader = o.ExternalOAuth[0].AuthorizationHeader
	}

	// If the state has no data for carts (0 items) and the configuration is the
	// default we match the state
	if p.Carts[0].isDefault() && (len(o.Carts) == 0 || o.Carts[0].isDefault()) {
		p.Carts = o.Carts
	}

	// The commercetools default for delete_days_after_creation is 15, so if the
	if len(p.Messages) > 0 && len(o.Messages) > 0 {
		if p.Messages[0].DeleteDaysAfterCreation.ValueInt64() == DefaultDeleteDaysAfterCreation && o.Messages[0].DeleteDaysAfterCreation.IsNull() {
			p.Messages[0].DeleteDaysAfterCreation = o.Messages[0].DeleteDaysAfterCreation
		}
	}
	// If the state has no data for messages (0 items) and the configuration is
	// the default we match the state
	if len(p.Messages) > 0 && p.Messages[0].isDefault() && (len(o.Messages) == 0 || o.Messages[0].isDefault()) {
		p.Messages = o.Messages
	}
}

func (p Project) updateActions(plan Project) platform.ProjectUpdate {
	result := platform.ProjectUpdate{
		Version: int(p.Version.ValueInt64()),
		Actions: []platform.ProjectUpdateAction{},
	}

	// changeMyBusinessUnitStatusOnCreation
	// TODO

	// changeCartsConfiguration
	if !reflect.DeepEqual(p.Carts, plan.Carts) {
		if len(plan.Carts) == 0 {
			result.Actions = append(result.Actions,
				platform.ProjectChangeCartsConfigurationAction{
					CartsConfiguration: platform.CartsConfiguration{},
				},
			)
		} else {
			val := plan.Carts[0].toNative()
			result.Actions = append(result.Actions,
				platform.ProjectChangeCartsConfigurationAction{
					CartsConfiguration: val,
				},
			)

			//ProjectChangeCartsConfigurationAction does not actually update CountryTaxRateFallbackEnabled,
			// so added extra mutation in same flow to keep consistent with previous code
			result.Actions = append(result.Actions,
				platform.ProjectChangeCountryTaxRateFallbackEnabledAction{
					CountryTaxRateFallbackEnabled: *val.CountryTaxRateFallbackEnabled,
				},
			)
		}
	}

	// changeCountries
	if !reflect.DeepEqual(p.Countries, plan.Countries) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeCountriesAction{
				Countries: pie.Map(plan.Countries, func(val types.String) string {
					return val.ValueString()
				}),
			},
		)
	}

	// changeCurrencies
	if !reflect.DeepEqual(p.Currencies, plan.Currencies) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeCurrenciesAction{
				Currencies: pie.Map(plan.Currencies, func(val types.String) string {
					return val.ValueString()
				}),
			},
		)
	}

	// changeLanguages
	if !reflect.DeepEqual(p.Languages, plan.Languages) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeLanguagesAction{
				Languages: pie.Map(plan.Languages, func(val types.String) string {
					return val.ValueString()
				}),
			},
		)
	}

	// changeMessagesConfiguration
	if !reflect.DeepEqual(p.Messages, plan.Messages) {
		if len(plan.Messages) > 0 {
			result.Actions = append(result.Actions,
				platform.ProjectChangeMessagesConfigurationAction{
					MessagesConfiguration: plan.Messages[0].toNative(),
				},
			)
		} else {
			// Set message configuration to the default values
			result.Actions = append(result.Actions,
				platform.ProjectChangeMessagesConfigurationAction{
					MessagesConfiguration: platform.MessagesConfigurationDraft{
						Enabled:                 false,
						DeleteDaysAfterCreation: DefaultDeleteDaysAfterCreation,
					},
				},
			)
		}
	}

	// changeName
	if !p.Name.Equal(plan.Name) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeNameAction{
				Name: plan.Name.ValueString(),
			},
		)
	}

	// changeOrderSearchStatus
	if !(p.EnableSearchIndexOrders.ValueBool() == plan.EnableSearchIndexOrders.ValueBool()) {
		status := platform.OrderSearchStatusDeactivated
		if plan.EnableSearchIndexOrders.ValueBool() {
			status = platform.OrderSearchStatusActivated
		}
		result.Actions = append(result.Actions,
			platform.ProjectChangeOrderSearchStatusAction{
				Status: status,
			},
		)
	}

	// changeProductSearchIndexingEnabled
	if !(p.EnableSearchIndexProducts.ValueBool() == plan.EnableSearchIndexProducts.ValueBool()) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeProductSearchIndexingEnabledAction{
				Enabled: plan.EnableSearchIndexProducts.ValueBool(),
			},
		)
	}

	// changeShoppingListsConfiguration
	// TODO

	// setExternalOAuth
	if !reflect.DeepEqual(p.ExternalOAuth, plan.ExternalOAuth) {
		var value *platform.ExternalOAuth
		if len(plan.ExternalOAuth) > 0 {
			value = plan.ExternalOAuth[0].toNative()
		}
		result.Actions = append(result.Actions,
			platform.ProjectSetExternalOAuthAction{
				ExternalOAuth: value,
			},
		)
	}

	// setShippingRateInputType
	if !p.ShippingRateInputType.Equal(plan.ShippingRateInputType) ||
		!reflect.DeepEqual(p.ShippingRateCartClassificationValue, plan.ShippingRateCartClassificationValue) {
		var value platform.ShippingRateInputType
		switch plan.ShippingRateInputType.ValueString() {
		case "CartClassification":
			value = platform.CartClassificationType{
				Values: pie.Map(
					plan.ShippingRateCartClassificationValue,
					func(v models.CustomFieldLocalizedEnumValue) platform.CustomFieldLocalizedEnumValue {
						return v.ToNative()
					}),
			}
		case "CartScore":
			value = platform.CartScoreType{}
		case "CartValue":
			value = platform.CartValueType{}
		}

		result.Actions = append(result.Actions,
			platform.ProjectSetShippingRateInputTypeAction{
				ShippingRateInputType: value,
			},
		)
	}

	return result
}

type Messages struct {
	Enabled                 types.Bool  `tfsdk:"enabled"`
	DeleteDaysAfterCreation types.Int64 `tfsdk:"delete_days_after_creation"`
}

func (m Messages) toNative() platform.MessagesConfigurationDraft {
	days := DefaultDeleteDaysAfterCreation // Commercetools default

	if !m.DeleteDaysAfterCreation.IsNull() {
		days = int(m.DeleteDaysAfterCreation.ValueInt64())
	}

	return platform.MessagesConfigurationDraft{
		Enabled:                 m.Enabled.ValueBool(),
		DeleteDaysAfterCreation: days,
	}
}
func (m Messages) isDefault() bool {
	return !m.toNative().Enabled && m.DeleteDaysAfterCreation.ValueInt64() == DefaultDeleteDaysAfterCreation
}

type ExternalOAuth struct {
	URL                 types.String `tfsdk:"url"`
	AuthorizationHeader types.String `tfsdk:"authorization_header"`
}

func (e ExternalOAuth) toNative() *platform.ExternalOAuth {
	return &platform.ExternalOAuth{
		Url:                 e.URL.ValueString(),
		AuthorizationHeader: e.AuthorizationHeader.ValueString(),
	}
}

type Carts struct {
	CountryTaxRateFallbackEnabled   types.Bool  `tfsdk:"country_tax_rate_fallback_enabled"`
	DeleteDaysAfterLastModification types.Int64 `tfsdk:"delete_days_after_last_modification"`
}

func (c Carts) isDefault() bool {
	return !c.CountryTaxRateFallbackEnabled.ValueBool() &&
		c.DeleteDaysAfterLastModification.IsNull()
}

func (c Carts) toNative() platform.CartsConfiguration {
	return platform.CartsConfiguration{
		DeleteDaysAfterLastModification: utils.OptionalInt(c.DeleteDaysAfterLastModification),
		CountryTaxRateFallbackEnabled:   utils.BoolRef(c.CountryTaxRateFallbackEnabled.ValueBool()),
	}
}
