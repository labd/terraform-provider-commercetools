package commercetools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/labd/commercetools-go-sdk/commercetools"
)

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

func stringFormatErrorExtras(err commercetools.Error) string {
	switch len(err.Errors) {
	case 0:
		return ""
	case 1:
		return stringFormatObject(err.Errors[0].Extra())
	default:
		{
			messages := make([]string, len(err.Errors))
			for i, item := range err.Errors {
				messages[i] = fmt.Sprintf(" %d. %s", i+1, stringFormatObject(item.Extra()))
			}
			return strings.Join(messages, "\n")
		}
	}
}

func stringFormatActions(actions commercetools.UpdateActions) string {
	lines := make([]string, len(actions))
	for i, action := range actions {
		lines[i] = fmt.Sprintf("%d: %s", i, stringFormatObject(action))
	}
	return strings.Join(lines, "\n")
}
