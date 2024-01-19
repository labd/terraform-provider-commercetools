package commercetools

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/labd/commercetools-go-sdk/platform"
)

func TestCreateLookup(t *testing.T) {
	input := []any{
		map[string]any{
			"name":  "name1",
			"value": "Value 1",
		},
		map[string]any{
			"name":  "name2",
			"value": "Value 2",
		},
	}
	result := createLookup(input, "name")
	if _, ok := result["name1"]; !ok {
		t.Error("Could not lookup name1")
	}
	if _, ok := result["name2"]; !ok {
		t.Error("Could not lookup name1")
	}
}

func TestCompareDateString(t *testing.T) {
	type testCase struct {
		a        string
		b        string
		expected bool
	}

	testCases := []testCase{
		{"2018-01-02T15:04:05.000Z", "2018-01-02T15:04:05.000Z", true},
		{"2017-03-04T10:01:02.000Z", "2018-01-02T15:04:05.000Z", false},
		{"2018-01-02T15:04:05.000Z", "2018-01-02T15:04:05Z", true},
		{"2018-01-02T15:04:05Z", "2018-01-02T15:04:05Z", true},
		{"2018-01-02T15:04:05Z", "2018-01-02T15:04:05.999Z", true},
		{"2018-01-02T15:04:04Z", "2018-01-02T15:04:05Z", false},
		{"2018-01-02T15:06:04Z", "2018-01-02T15:04:05.999Z", false},
		{"", "2018-01-02T15:04:05.999Z", false},
		{"", "xxx", false},
		{"", "", true},
	}

	var res bool
	for _, tt := range testCases {
		res = compareDateString(tt.a, tt.b)
		if res != tt.expected {
			t.Errorf("failures %v, got %v", tt.expected, res)
		}
	}

}

func checkApiResult(err error) error {
	if errors.Is(err, platform.ErrNotFound) {
		return nil
	}

	switch v := err.(type) {
	case platform.GenericRequestError:
		if v.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("unhandled error generic error returned (%d)", v.StatusCode)
	case platform.ResourceNotFoundError:
		return nil
	default:
		return fmt.Errorf("unexpected result returned")
	}
}

func TestIntNilIfEmpty(t *testing.T) {
	testCases := []struct {
		input    *int
		expected *int
	}{
		{nil, nil},
		{intRef(0), nil},
		{intRef(1), intRef(1)},
	}

	for _, tt := range testCases {
		v := intNilIfEmpty(tt.input)
		assert.Equal(t, tt.expected, v)
	}
}

func TestValidateLocalizedStringKey(t *testing.T) {
	testCases := []struct {
		input    interface{}
		failures int
	}{
		{map[string]any{"en": "English"}, 0},
		{map[string]any{"en-US": "English (United States)"}, 0},
		{map[string]any{"es-419": "Spanish (Latin America)"}, 0},
		{map[string]any{"rm-sursilv": "Romansh Sursilvan"}, 0},
		{map[string]any{"sr-Cyrl": "Serbian in Cyrillic"}, 0},
		{map[string]any{"foobar": "Fail"}, 1},
		{"foobar", 1},
		{1, 1},
		{map[string]any{"es-409": "Spanish (Latin America)"}, 1},
	}

	for _, tt := range testCases {
		diag := validateLocalizedStringKey(tt.input, nil)
		assert.Equal(t, tt.failures, len(diag), fmt.Sprintf("%+v", diag))
	}
}
