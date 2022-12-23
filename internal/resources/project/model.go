package project

import (
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/models"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
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

	Carts         *Carts         `tfsdk:"carts"`
	Messages      *Messages      `tfsdk:"messages"`
	ExternalOAuth *ExternalOAuth `tfsdk:"external_oauth"`

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

		/* Carts: &Carts{ */
		/* 	DeleteDaysAfterLastModification: utils.FromOptionalInt(n.Carts.DeleteDaysAfterLastModification), */
		/* 	CountryTaxRateFallbackEnabled:   types.BoolValue(*n.Carts.CountryTaxRateFallbackEnabled), */
		/* }, */
		/* Messages: &Messages{ */
		/* 	DeleteDaysAfterCreation: utils.FromOptionalInt(n.Messages.DeleteDaysAfterCreation), */
		/* 	Enabled:                 types.BoolValue(n.Messages.Enabled), */
		/* }, */
		/* ExternalOAuth: nil, */
	}

	switch s := n.ShippingRateInputType.(type) {
	case platform.CartScoreType:
		res.ShippingRateInputType = types.StringValue("CartScore")
		res.ShippingRateCartClassificationValue = []models.CustomFieldLocalizedEnumValue{}
	case platform.CartValueType:
		res.ShippingRateInputType = types.StringValue("CartValue")
		res.ShippingRateCartClassificationValue = []models.CustomFieldLocalizedEnumValue{}
	case platform.CartClassificationType:
		res.ShippingRateInputType = types.StringValue("CartClassification")
		values := make([]models.CustomFieldLocalizedEnumValue, len(s.Values))
		for i := range s.Values {
			values[i] = models.NewCustomFieldLocalizedEnumValue(s.Values[i])
		}
		res.ShippingRateCartClassificationValue = values
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

	/* if n.ExternalOAuth != nil { */
	/* 	res.ExternalOAuth = &ExternalOAuth{ */
	/* 		URL:                 types.StringValue(n.ExternalOAuth.Url), */
	/* 		AuthorizationHeader: types.StringUnknown(), */
	/* 	} */
	/* } */

	return res
}

func (p *Project) SetNewData(o Project) {
	p.Version = o.Version

	/* if p.Carts != nil && o.Carts != nil { */
	/* 	p.Carts.DeleteDaysAfterLastModification = o.Carts.DeleteDaysAfterLastModification */
	/* } */

}

func (p *Project) SetStateData(o Project) {
	/*
		if p.ExternalOAuth != nil {
			p.ExternalOAuth.AuthorizationHeader = o.ExternalOAuth.AuthorizationHeader
		}

		// If the state has no data for carts (nil) and the configuration is the
		// default we match the state
		if o.Carts == nil &&
			!p.Carts.CountryTaxRateFallbackEnabled.ValueBool() &&
			p.Carts.DeleteDaysAfterLastModification.IsNull() {
			p.Carts = nil
		}
		if o.Carts != nil && o.Carts.CountryTaxRateFallbackEnabled.IsNull() &&
			p.Carts != nil && !p.Carts.CountryTaxRateFallbackEnabled.ValueBool() {
			p.Carts.CountryTaxRateFallbackEnabled = types.BoolNull()
		}
	*/
}

func (p Project) UpdateActions(n Project) platform.ProjectUpdate {
	result := platform.ProjectUpdate{
		Version: int(p.Version.ValueInt64()),
		Actions: []platform.ProjectUpdateAction{},
	}

	if !p.Name.Equal(n.Name) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeNameAction{
				Name: n.Name.ValueString(),
			},
		)
	}

	if !reflect.DeepEqual(p.Countries, n.Countries) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeCountriesAction{
				Countries: pie.Map(n.Countries, func(val types.String) string {
					return val.ValueString()
				}),
			},
		)
	}

	if !reflect.DeepEqual(p.Currencies, n.Currencies) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeCurrenciesAction{
				Currencies: pie.Map(n.Currencies, func(val types.String) string {
					return val.ValueString()
				}),
			},
		)
	}

	if !reflect.DeepEqual(p.Languages, n.Languages) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeLanguagesAction{
				Languages: pie.Map(n.Languages, func(val types.String) string {
					return val.ValueString()
				}),
			},
		)
	}

	if !p.EnableSearchIndexProducts.Equal(n.EnableSearchIndexProducts) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeProductSearchIndexingEnabledAction{
				Enabled: n.EnableSearchIndexProducts.ValueBool(),
			},
		)
	}

	if !p.EnableSearchIndexOrders.Equal(n.EnableSearchIndexOrders) {
		status := platform.OrderSearchStatusDeactivated
		if n.EnableSearchIndexOrders.ValueBool() {
			status = platform.OrderSearchStatusActivated
		}
		result.Actions = append(result.Actions,
			platform.ProjectChangeOrderSearchStatusAction{
				Status: status,
			},
		)
	}

	if !p.ShippingRateInputType.Equal(n.ShippingRateInputType) {
		var value platform.ShippingRateInputType
		switch n.ShippingRateInputType.ValueString() {
		case "CartClassification":
			value = platform.CartClassificationType{
				Values: pie.Map(
					n.ShippingRateCartClassificationValue,
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

	if !reflect.DeepEqual(p.Messages, n.Messages) {
		result.Actions = append(result.Actions,
			platform.ProjectChangeMessagesConfigurationAction{
				MessagesConfiguration: n.Messages.toNative(),
			},
		)
	}

	/* if !reflect.DeepEqual(p.Carts, n.Carts) { */
	/* 	log.Println(spew.Sdump(p.Carts)) */
	/* 	log.Println(spew.Sdump(n.Carts)) */

	/* 	if (p.Carts != nil && !p.Carts.isDefault()) || */
	/* 		(n.Carts != nil && !n.Carts.isDefault()) { */

	/* 		var val platform.CartsConfiguration */
	/* 		if n.Carts != nil { */
	/* 			val = n.Carts.toNative() */
	/* 		} */
	/* 		result.Actions = append(result.Actions, */
	/* 			platform.ProjectChangeCartsConfigurationAction{ */
	/* 				CartsConfiguration: val, */
	/* 			}, */
	/* 		) */
	/* 	} */
	/* } */

	/* if !reflect.DeepEqual(p.ExternalOAuth, n.ExternalOAuth) { */
	/* 	var value *platform.ExternalOAuth */
	/* 	if n.ExternalOAuth != nil { */
	/* 		value = n.ExternalOAuth.toNative() */

	/* 	} */
	/* 	result.Actions = append(result.Actions, */
	/* 		platform.ProjectSetExternalOAuthAction{ */
	/* 			ExternalOAuth: value, */
	/* 		}, */
	/* 	) */
	/* } */

	return result
}

type Messages struct {
	Enabled                 types.Bool  `tfsdk:"enabled"`
	DeleteDaysAfterCreation types.Int64 `tfsdk:"delete_days_after_creation"`
}

func (m Messages) toNative() platform.MessagesConfigurationDraft {
	return platform.MessagesConfigurationDraft{
		Enabled:                 m.Enabled.ValueBool(),
		DeleteDaysAfterCreation: int(m.DeleteDaysAfterCreation.ValueInt64()),
	}
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
