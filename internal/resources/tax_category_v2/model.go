package tax_category_v2

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"reflect"
)

type TaxCategory struct {
	ID          types.String `tfsdk:"id"`
	Version     types.Int64  `tfsdk:"version"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	TaxRates    []TaxRate    `tfsdk:"tax_rates"`
}

type TaxRate struct {
	ID              types.String  `tfsdk:"id"`
	Key             types.String  `tfsdk:"key"`
	Name            types.String  `tfsdk:"name"`
	Amount          types.Float64 `tfsdk:"amount"`
	IncludedInPrice types.Bool    `tfsdk:"included_in_price"`
	Country         types.String  `tfsdk:"country"`
	State           types.String  `tfsdk:"state"`
	SubRates        []SubRate     `tfsdk:"sub_rates"`
}

type SubRate struct {
	Name   types.String  `tfsdk:"name"`
	Amount types.Float64 `tfsdk:"amount"`
}

func (tr TaxRate) draft() platform.TaxRateDraft {
	var subRateDrafts = make([]platform.SubRate, 0, len(tr.SubRates))

	for _, subRate := range tr.SubRates {
		subRateDrafts = append(subRateDrafts, platform.SubRate{
			Name:   subRate.Name.ValueString(),
			Amount: subRate.Amount.ValueFloat64(),
		})
	}

	return platform.TaxRateDraft{
		Name:            tr.Name.ValueString(),
		Amount:          tr.Amount.ValueFloat64Pointer(),
		IncludedInPrice: tr.IncludedInPrice.ValueBool(),
		Country:         tr.Country.ValueString(),
		State:           tr.State.ValueStringPointer(),
		SubRates:        subRateDrafts,
	}
}

func (tc TaxCategory) draft() platform.TaxCategoryDraft {
	var taxRateDrafts = make([]platform.TaxRateDraft, 0, len(tc.TaxRates))

	for _, taxRate := range tc.TaxRates {
		taxRateDrafts = append(taxRateDrafts, taxRate.draft())
	}

	return platform.TaxCategoryDraft{
		Name:        tc.Name.ValueString(),
		Key:         tc.Key.ValueStringPointer(),
		Description: tc.Description.ValueStringPointer(),
		Rates:       taxRateDrafts,
	}
}

func TaxCategoryFromNative(tc *platform.TaxCategory) (TaxCategory, error) {
	var rates = make([]TaxRate, 0, len(tc.Rates))
	for _, taxRate := range tc.Rates {
		var idPtr = taxRate.ID
		if idPtr == nil {
			return TaxCategory{}, fmt.Errorf("tax rate ID is nil")
		}

		var subRates = make([]SubRate, 0, len(taxRate.SubRates))
		for _, subRate := range taxRate.SubRates {
			var subRate = SubRate{
				Name:   types.StringValue(subRate.Name),
				Amount: types.Float64Value(subRate.Amount),
			}
			subRates = append(subRates, subRate)
		}

		var rate = TaxRate{
			ID:              types.StringValue(*idPtr),
			Key:             types.StringPointerValue(taxRate.Key),
			Name:            types.StringValue(taxRate.Name),
			Amount:          types.Float64Value(taxRate.Amount),
			IncludedInPrice: types.BoolValue(taxRate.IncludedInPrice),
			Country:         types.StringValue(taxRate.Country),
			State:           types.StringPointerValue(taxRate.State),
			SubRates:        subRates,
		}
		rates = append(rates, rate)
	}

	return TaxCategory{
		ID:          types.StringValue(tc.ID),
		Version:     types.Int64Value(int64(tc.Version)),
		Key:         types.StringPointerValue(tc.Key),
		Name:        types.StringValue(tc.Name),
		Description: types.StringPointerValue(tc.Description),
		TaxRates:    rates,
	}, nil
}

func (tc TaxCategory) updateActions(plan TaxCategory) platform.TaxCategoryUpdate {
	result := platform.TaxCategoryUpdate{
		Version: int(tc.Version.ValueInt64()),
		Actions: []platform.TaxCategoryUpdateAction{},
	}

	// setKey
	if !tc.Key.Equal(plan.Key) {
		var newKey *string
		if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
			newKey = plan.Key.ValueStringPointer()
		}

		result.Actions = append(
			result.Actions,
			platform.TaxCategorySetKeyAction{Key: newKey},
		)
	}

	// changeName
	if !tc.Name.Equal(plan.Name) {
		var newName string
		if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
			newName = plan.Name.ValueString()
		}

		result.Actions = append(
			result.Actions,
			platform.TaxCategoryChangeNameAction{Name: newName},
		)
	}

	// setDescription
	if !tc.Description.Equal(plan.Description) {
		var newDescription *string
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			newDescription = plan.Description.ValueStringPointer()
		}

		result.Actions = append(
			result.Actions,
			platform.TaxCategorySetDescriptionAction{Description: newDescription},
		)
	}

	// If a change occurred in tax rates we remove all the tax rates and re-add them.
	// This is because a change manual change in tax rate through the API or merchant
	// center re-creates the tax rate so the ID cannot be trusted
	if !reflect.DeepEqual(tc.TaxRates, plan.TaxRates) {
		for _, tr := range tc.TaxRates {
			result.Actions = append(
				result.Actions,
				platform.TaxCategoryRemoveTaxRateAction{TaxRateId: tr.ID.ValueStringPointer()},
			)
		}

		for _, dtr := range plan.TaxRates {
			result.Actions = append(
				result.Actions,
				platform.TaxCategoryAddTaxRateAction{TaxRate: dtr.draft()})
		}
	}

	return result
}
