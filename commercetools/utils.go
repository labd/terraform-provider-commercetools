package commercetools

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

// TypeLocalizedString defined merely for documentation,
// it basically is just a normal TypeMap but clearifies in the code that
// it should be used to store a LocalizedString
const TypeLocalizedString = schema.TypeMap

func getClient(m interface{}) *platform.ByProjectKeyRequestBuilder {
	client := m.(*platform.ByProjectKeyRequestBuilder)
	return client
}

func stringRef(value interface{}) *string {
	result := value.(string)
	return &result
}

func intRef(value interface{}) *int {
	result := value.(int)
	return &result
}

func boolRef(value interface{}) *bool {
	result := value.(bool)
	return &result
}

func handleCommercetoolsError(err error) *resource.RetryError {
	if ctErr, ok := err.(platform.ErrorResponse); ok {
		return resource.NonRetryableError(ctErr)
	}

	log.Printf("[DEBUG] Received error: %s", err)
	return resource.RetryableError(err)
}

func expandStringArray(input []interface{}) []string {
	s := make([]string, len(input))
	for i, v := range input {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func expandStringMap(input map[string]interface{}) map[string]string {
	s := make(map[string]string)
	for k, v := range input {
		s[k] = fmt.Sprint(v)
	}
	return s
}

func localizedStringCompare(a platform.LocalizedString, b map[string]interface{}) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func localizedStringToMap(input platform.LocalizedString) map[string]string {
	result := make(map[string]string, len(input))
	for k, v := range input {
		result[k] = v
	}
	return result
}

func stringFormatObject(object interface{}) string {
	data, err := json.MarshalIndent(object, "", "    ")

	if err != nil {
		return fmt.Sprintf("%+v", object)
	}
	return string(append(data, '\n'))
}

func stringFormatErrorExtras(err platform.ErrorResponse) string {
	switch len(err.Errors) {
	case 0:
		return ""
	case 1:
		return "temp" // stringFormatObject(err.Errors[0].Error())
	default:
		{
			messages := make([]string, len(err.Errors))
			for i, item := range err.Errors {
				messages[i] = fmt.Sprintf("%v", item)
				//messages[i] = fmt.Sprintf(" %d. %s", i+1, stringFormatObject(item.Extra()))
			}
			return strings.Join(messages, "\n")
		}
	}
}

func stringFormatActions(actions ...interface{}) string {
	lines := []string{}
	for i, action := range actions {
		lines = append(
			lines,
			fmt.Sprintf("%d: %s", i, stringFormatObject(action)))

	}
	return strings.Join(lines, "\n")
}

func createLookup(objects []interface{}, key string) map[string]interface{} {
	lookup := make(map[string]interface{})
	for _, field := range objects {
		f := field.(map[string]interface{})
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
func ValidateCurrencyCode(val interface{}, key string) (warns []string, errs []error) {
	currency := val.(string)
	if _, exists := currencyCodes[currency]; !exists {
		errs = append(errs, fmt.Errorf("%q unknown currency code, must be valid ISO 4217 code, got: %s", key, currency))
	}
	return
}

func expandDate(input string) (time.Time, error) {
	return time.Parse(time.RFC3339, input)
}
