package commercetools

import (
	"time"

	"github.com/labd/commercetools-go-sdk/platform"
)

func marshallTime(val *time.Time) string {
	if val == nil {
		return ""
	}
	return val.Format(time.RFC3339)
}

func unmarshallTime(input string) (time.Time, error) {
	return time.Parse(time.RFC3339, input)
}

func marshallTypedMoney(val platform.TypedMoney) map[string]interface{} {
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
	}
	panic("Unknown money type")
}

func unmarshallTypedMoney(d map[string]interface{}) []platform.Money {
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
