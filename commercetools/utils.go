package commercetools

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/labd/commercetools-go-sdk/platform"
)

// TypeLocalizedString defined merely for documentation,
// it basically is just a normal TypeMap but clearifies in the code that
// it should be used to store a LocalizedString
const TypeLocalizedString = schema.TypeMap

func getClient(m any) *platform.ByProjectKeyRequestBuilder {
	client := m.(*platform.ByProjectKeyRequestBuilder)
	return client
}

func stringRef(value any) *string {
	if _, ok := value.(*string); ok {
		return value.(*string)
	}

	if value == nil {
		return nil
	}
	result := value.(string)
	return &result
}

func stringUnref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intRef(value any) *int {
	result := value.(int)
	return &result
}

func boolRef(value any) *bool {
	result := value.(bool)
	return &result
}

func expandStringArray(input []any) []string {
	s := make([]string, len(input))
	for i := range input {
		s[i] = input[i].(string)
	}
	return s
}

func createLookup(objects []any, key string) map[string]any {
	lookup := make(map[string]any)
	for _, field := range objects {
		f := field.(map[string]any)
		lookup[f[key].(string)] = field
	}
	return lookup
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

var currencyCodes = map[string]bool{
	"AED": true,
	"AFN": true,
	"ALL": true,
	"AMD": true,
	"ANG": true,
	"AOA": true,
	"ARS": true,
	"AUD": true,
	"AWG": true,
	"AZN": true,
	"BAM": true,
	"BBD": true,
	"BDT": true,
	"BGN": true,
	"BHD": true,
	"BIF": true,
	"BMD": true,
	"BND": true,
	"BOB": true,
	"BOV": true,
	"BRL": true,
	"BSD": true,
	"BTN": true,
	"BWP": true,
	"BYN": true,
	"BZD": true,
	"CAD": true,
	"CDF": true,
	"CHE": true,
	"CHF": true,
	"CHW": true,
	"CLF": true,
	"CLP": true,
	"CNY": true,
	"COP": true,
	"COU": true,
	"CRC": true,
	"CUC": true,
	"CUP": true,
	"CVE": true,
	"CZK": true,
	"DJF": true,
	"DKK": true,
	"DOP": true,
	"DZD": true,
	"EGP": true,
	"ERN": true,
	"ETB": true,
	"EUR": true,
	"FJD": true,
	"FKP": true,
	"GBP": true,
	"GEL": true,
	"GHS": true,
	"GIP": true,
	"GMD": true,
	"GNF": true,
	"GTQ": true,
	"GYD": true,
	"HKD": true,
	"HNL": true,
	"HRK": true,
	"HTG": true,
	"HUF": true,
	"IDR": true,
	"ILS": true,
	"INR": true,
	"IQD": true,
	"IRR": true,
	"ISK": true,
	"JMD": true,
	"JOD": true,
	"JPY": true,
	"KES": true,
	"KGS": true,
	"KHR": true,
	"KMF": true,
	"KPW": true,
	"KRW": true,
	"KWD": true,
	"KYD": true,
	"KZT": true,
	"LAK": true,
	"LBP": true,
	"LKR": true,
	"LRD": true,
	"LSL": true,
	"LYD": true,
	"MAD": true,
	"MDL": true,
	"MGA": true,
	"MKD": true,
	"MMK": true,
	"MNT": true,
	"MOP": true,
	"MRU": true,
	"MUR": true,
	"MVR": true,
	"MWK": true,
	"MXN": true,
	"MXV": true,
	"MYR": true,
	"MZN": true,
	"NAD": true,
	"NGN": true,
	"NIO": true,
	"NOK": true,
	"NPR": true,
	"NZD": true,
	"OMR": true,
	"PAB": true,
	"PEN": true,
	"PGK": true,
	"PHP": true,
	"PKR": true,
	"PLN": true,
	"PYG": true,
	"QAR": true,
	"RON": true,
	"RSD": true,
	"RUB": true,
	"RWF": true,
	"SAR": true,
	"SBD": true,
	"SCR": true,
	"SDG": true,
	"SEK": true,
	"SGD": true,
	"SHP": true,
	"SLL": true,
	"SOS": true,
	"SRD": true,
	"SSP": true,
	"STN": true,
	"SVC": true,
	"SYP": true,
	"SZL": true,
	"THB": true,
	"TJS": true,
	"TMT": true,
	"TND": true,
	"TOP": true,
	"TRY": true,
	"TTD": true,
	"TWD": true,
	"TZS": true,
	"UAH": true,
	"UGX": true,
	"USD": true,
	"USN": true,
	"UYI": true,
	"UYU": true,
	"UZS": true,
	"VEF": true,
	"VND": true,
	"VUV": true,
	"WST": true,
	"XAF": true,
	"XAG": true,
	"XAU": true,
	"XBA": true,
	"XBB": true,
	"XBC": true,
	"XBD": true,
	"XCD": true,
	"XDR": true,
	"XOF": true,
	"XPD": true,
	"XPF": true,
	"XPT": true,
	"XSU": true,
	"XTS": true,
	"XUA": true,
	"YER": true,
	"ZAR": true,
	"ZMW": true,
	"ZWL": true,
}

// ValidateCurrencyCode checks if a currency string is valid according to https://en.wikipedia.org/wiki/ISO_4217
func ValidateCurrencyCode(val any, key string) (warns []string, errs []error) {
	currency := val.(string)
	if _, exists := currencyCodes[currency]; !exists {
		errs = append(errs, fmt.Errorf("%q unknown currency code, must be valid ISO 4217 code, got: %s", key, currency))
	}
	return
}

func transformToList(data map[string]any, key string) {
	newDestination := make([]any, 1)
	if data[key] != nil {
		newDestination[0] = data[key]
	}
	data[key] = newDestination
}

func elementFromList(d *schema.ResourceData, key string) (map[string]any, error) {
	data := d.Get(key).([]any)

	if len(data) > 0 {
		result := data[0].(map[string]any)
		return result, nil
	}
	return nil, nil
}

func firstElementFromSlice(d []any) map[string]any {
	if len(d) > 0 {
		result := d[0].(map[string]any)
		return result
	}
	return nil
}

func elementFromSlice(d map[string]any, key string) map[string]any {
	data, ok := d[key]
	if !ok {
		return nil
	}

	items := data.([]any)
	if len(items) > 0 {
		result := items[0].(map[string]any)
		return result
	}
	return nil
}

func isNotEmpty(d map[string]any, key string) (any, bool) {
	val, ok := d[key]
	if !ok {
		return nil, false
	}

	if val != "" {
		return val, true
	}
	return nil, false
}

// nilIfEmpty returns a nil value if the string is nil or empty ("") otherwise
// it returns the value
func nilIfEmpty(val *string) *string {
	if val == nil {
		return nil
	}

	if *val == "" {
		return nil
	}
	return val
}

var validateLocalizedStringKey = validation.MapKeyMatch(
	regexp.MustCompile("^[a-z]{2}(-[A-Z]{2})?$"),
	"Locale keys must match pattern ^[a-z]{2}(-[A-Z]{2})?$",
)

func upperStringSlice(items []string) []string {
	s := make([]string, len(items))
	for i, v := range items {
		s[i] = strings.ToUpper(v)
	}
	return s
}

// languageCode converts an IETF language tag with mixed casing into the case-sensitive format.
// The original item is returned if the given input is not valid.
func languageCode(s string) string {
	if len(s) == 2 {
		return strings.ToLower(s)
	}
	parts := strings.Split(s, "-")
	if len(parts) == 2 {
		return strings.Join([]string{strings.ToLower(parts[0]), strings.ToUpper(parts[1])}, "-")
	}
	// fallback to the original
	return s
}

func languageCodeSlice(items []string) []string {
	codes := make([]string, len(items))
	for i, code := range items {
		codes[i] = languageCode(code)
	}
	return codes
}

func compareDateString(a, b string) bool {
	if a == b {
		return true
	}
	da, err := expandTime(a)
	if err != nil {
		return false
	}
	db, err := expandTime(b)
	if err != nil {
		return false
	}
	return da.Unix() == db.Unix()
}

func diffSuppressDateString(k, old, new string, d *schema.ResourceData) bool {
	return compareDateString(old, new)
}

func removeValueFromSlice(items []string, value string) []string {
	for i, v := range items {
		if v == value {
			return append(items[:i], items[i+1:]...)
		}
	}
	return items
}

// diffSlices does a diff on two slices and returns the changes. If a field is
// no longer available then nil is returned.
func diffSlices(old, new map[string]any) map[string]any {
	result := map[string]any{}
	seen := map[string]bool{}

	// Find changes against current values. If value no longer
	// exists we set it to nil
	for key, value := range old {
		seen[key] = true
		newVal, exists := new[key]
		if !exists {
			result[key] = nil
			continue
		}

		if !reflect.DeepEqual(value, newVal) {
			result[key] = newVal
			continue
		}
	}

	// Copy new values
	for key, value := range new {
		if _, exists := seen[key]; !exists {
			result[key] = value
		}
	}

	return result
}

func coerceTypedMoney(val platform.TypedMoney) platform.Money {
	switch p := val.(type) {
	case platform.CentPrecisionMoney:
		return platform.Money{
			CentAmount:   p.CentAmount,
			CurrencyCode: p.CurrencyCode,
		}
	case platform.HighPrecisionMoney:
		return platform.Money{
			CentAmount:   p.CentAmount,
			CurrencyCode: p.CurrencyCode,
		}
	}

	return platform.Money{}
}
