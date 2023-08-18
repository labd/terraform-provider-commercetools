package commercetools

import (
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
	"testing"
)

var customFieldEncodeValueTests = []struct {
	typ         any
	value       any
	expectedVal any
	hasError    bool
}{
	//CustomFieldLocalizedStringType
	{typ: platform.CustomFieldLocalizedStringType{}, value: `{"foo":"bar"}`, expectedVal: platform.LocalizedString{"foo": "bar"}},
	{typ: platform.CustomFieldLocalizedStringType{}, value: `foobar`, hasError: true},

	//CustomFieldBooleanType
	{typ: platform.CustomFieldBooleanType{}, value: "true", expectedVal: true},
	{typ: platform.CustomFieldBooleanType{}, value: "false", expectedVal: false},
	{typ: platform.CustomFieldBooleanType{}, value: "foobar", hasError: true},

	//CustomFieldNumberType
	{typ: platform.CustomFieldNumberType{}, value: "1", expectedVal: int64(1)},
	{typ: platform.CustomFieldNumberType{}, value: "foobar", hasError: true},

	//CustomFieldSetType
	{
		typ:         platform.CustomFieldSetType{ElementType: platform.CustomFieldStringType{}},
		value:       `["hello", "world"]`,
		expectedVal: []interface{}{"hello", "world"},
	},
	{
		typ:         platform.CustomFieldSetType{ElementType: platform.CustomFieldNumberType{}},
		value:       `[1, 2]`,
		expectedVal: []interface{}{int64(1), int64(2)},
	},
	{
		typ:   platform.CustomFieldSetType{ElementType: platform.CustomFieldReferenceType{}},
		value: `[{"id":"98edd6e4-1702-45d5-8bc0-bbb792a4a839","typeId":"zone"},{"id":"8a8efb57-71d3-4a8d-aa77-4d4e6df9ef2a","typeId":"zone"}]`,
		expectedVal: []interface{}{
			map[string]interface{}{"id": "98edd6e4-1702-45d5-8bc0-bbb792a4a839", "typeId": "zone"},
			map[string]interface{}{"id": "8a8efb57-71d3-4a8d-aa77-4d4e6df9ef2a", "typeId": "zone"},
		},
	},

	//CustomFieldReferenceType
	{
		typ:         platform.CustomFieldReferenceType{},
		value:       `{"id":"98edd6e4-1702-45d5-8bc0-bbb792a4a839","typeId":"zone"}`,
		expectedVal: map[string]interface{}{"id": "98edd6e4-1702-45d5-8bc0-bbb792a4a839", "typeId": "zone"},
	},
}

func TestCustomFieldEncodeValue(t *testing.T) {
	for _, tt := range customFieldEncodeValueTests {
		t.Run("TestCustomFieldEncodeValue", func(t *testing.T) {
			encodedValue, err := customFieldEncodeValue(tt.typ, "some_field", tt.value)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.expectedVal, encodedValue)
		})
	}
}
