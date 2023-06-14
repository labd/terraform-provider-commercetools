package commercetools

import (
	"time"

	"fmt"
	"github.com/labd/commercetools-go-sdk/platform"
)

func flattenTime(val *time.Time) string {
	if val == nil {
		return ""
	}
	return val.Format(time.RFC3339)
}

func expandTime(input string) (time.Time, error) {
	return time.Parse(time.RFC3339, input)
}

func flattenTypedMoney(val platform.TypedMoney) map[string]any {
	switch v := val.(type) {
	case platform.HighPrecisionMoney:
		return map[string]any{
			"currency_code": v.CurrencyCode,
			"cent_amount":   v.CentAmount,
		}
	case platform.Money:
		return map[string]any{
			"currency_code": v.CurrencyCode,
			"cent_amount":   v.CentAmount,
		}
	case platform.CentPrecisionMoney:
		return map[string]any{
			"currency_code": v.CurrencyCode,
			"cent_amount":   v.CentAmount,
		}
	}
	panic(fmt.Sprintf("Unknown money type: %T", val))
}

func expandTypedMoney(d map[string]any) []platform.Money {
	input := d["money"].([]any)
	var result []platform.Money

	for _, raw := range input {
		i := raw.(map[string]any)
		priceCurrencyCode := i["currency_code"].(string)

		result = append(result, platform.Money{
			CurrencyCode: priceCurrencyCode,
			CentAmount:   i["cent_amount"].(int),
		})
	}

	return result
}

func expandLocalizedString(val any) platform.LocalizedString {
	values, ok := val.(map[string]any)
	if !ok {
		return platform.LocalizedString{}
	}

	result := make(platform.LocalizedString, len(values))
	for k := range values {
		result[k] = values[k].(string)
	}
	return result
}

func expandCentPrecisionMoneyDraft(d map[string]any) []platform.CentPrecisionMoneyDraft {
	input := d["money"].([]any)
	var result []platform.CentPrecisionMoneyDraft
	for _, raw := range input {
		data := raw.(map[string]any)
		item := platform.CentPrecisionMoneyDraft{}
		if currencyCode, ok := data["currency_code"].(string); ok {
			item.CurrencyCode = currencyCode
		}
		if centAmount, ok := data["cent_amount"].(int); ok {
			item.CentAmount = centAmount
		}
		if fractionDigits, ok := data["fraction_digits"].(int); ok {
			item.FractionDigits = &fractionDigits
		}
		result = append(result, item)
	}
	return result
}

func expandMoneyDraft(d map[string]any) []platform.Money {
	input := d["money"].([]any)
	var result []platform.Money
	for _, raw := range input {
		data := raw.(map[string]any)
		item := platform.Money{}
		if currencyCode, ok := data["currency_code"].(string); ok {
			item.CurrencyCode = currencyCode
		}
		if centAmount, ok := data["cent_amount"].(int); ok {
			item.CentAmount = centAmount
		}
		result = append(result, item)
	}
	return result
}
