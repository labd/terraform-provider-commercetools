package commercetools

import (
	"encoding/json"
	"fmt"
)

func expandStringArray(input []interface{}) []string {
	s := make([]string, len(input))
	for i, v := range input {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func stringFormatObject(object interface{}) string {
	data, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", object)
	}
	return string(append(data, '\n'))
}
