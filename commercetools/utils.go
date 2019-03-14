package commercetools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

// TypeLocalizedString defined merely for documentation,
// it basically is just a normal TypeMap but clearifies in the code that
// it should be used to store a LocalizedString
const TypeLocalizedString = schema.TypeMap

func getClient(m interface{}) *commercetools.Client {
	client := m.(*commercetools.Client)
	return client
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

func localizedStringToMap(input commercetools.LocalizedString) map[string]interface{} {
	result := make(map[string]interface{}, len(input))
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

func stringFormatErrorExtras(err commercetools.ErrorResponse) string {
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
