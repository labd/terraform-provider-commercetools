package commercetools

import (
	"time"

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

func flattenTypedMoney(val platform.TypedMoney) map[string]interface{} {
	switch v := val.(type) {
	case platform.HighPrecisionMoney:
		return map[string]interface{}{
			"currency_code": v.CurrencyCode,
			"cent_amount":   v.CentAmount,
		}
	case platform.Money:
		return map[string]interface{}{
			"currency_code": v.CurrencyCode,
			"cent_amount":   v.CentAmount,
		}
	case platform.CentPrecisionMoney:
		return map[string]interface{}{
			"currency_code": v.CurrencyCode,
			"cent_amount":   v.CentAmount,
		}
	}
	panic("Unknown money type")
}

func expandTypedMoney(d map[string]interface{}) []platform.Money {
	input := d["money"].([]interface{})
	var result []platform.Money

	for _, raw := range input {
		i := raw.(map[string]interface{})
		priceCurrencyCode := i["currency_code"].(string)

		result = append(result, platform.Money{
			CurrencyCode: priceCurrencyCode,
			CentAmount:   i["cent_amount"].(int),
		})
	}

	return result
}

func expandLocalizedString(val interface{}) platform.LocalizedString {
	values, ok := val.(map[string]interface{})
	if !ok {
		return platform.LocalizedString{}
	}

	result := make(platform.LocalizedString, len(values))
	for k := range values {
		result[k] = values[k].(string)
	}
	return result
}

func expandCentPrecisionMoneyDraft(d map[string]interface{}) []platform.CentPrecisionMoneyDraft {
	input := d["money"].([]interface{})
	var result []platform.CentPrecisionMoneyDraft
	for _, raw := range input {
		data := raw.(map[string]interface{})
		item := platform.CentPrecisionMoneyDraft{}
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
