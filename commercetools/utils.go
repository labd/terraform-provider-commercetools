package commercetools

import "fmt"

func expandStringArray(input []interface{}) []string {
	s := make([]string, len(input))
	for i, v := range input {
		s[i] = fmt.Sprint(v)
	}
	return s
}
