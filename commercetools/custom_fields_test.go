package commercetools

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
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

// List of the resources with custom fields support
var customFieldResourceTypes = []string{"commercetools_channel", "commercetools_cart_discount", "commercetools_category",
	"commercetools_customer_group", "commercetools_discount_code", "commercetools_shipping_method", "commercetools_store"}

// List of the custom field types
var customFieldTypes = []string{"String", "Boolean", "Number", "LocalizedString", "Enum", "LocalizedEnum", "Money",
	"Date", "Time", "DateTime", "Reference", "Set"}

func TestAccCustomField_SetAndRemove(t *testing.T) {
	for _, customFieldResourceType := range customFieldResourceTypes {
		fmt.Println("Testing custom fields for:", customFieldResourceType)
		resourceShortName := "ct" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		resourceFullName := customFieldResourceType + "." + resourceShortName
		resourceKey := "key" + resourceShortName

		// Define Test Steps
		var customFieldsAccTestSteps = []resource.TestStep{
			{
				Config: getResourceConfig(customFieldResourceType, resourceShortName, resourceKey, customFieldTypes),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						customFields, err := getResourceCustomFields(s, customFieldResourceType, resourceFullName)
						if err != nil {
							return err
						}
						productType, err := testGetProductType(s, "commercetools_product_type.test")
						if err != nil {
							return err
						}
						assert.EqualValues(t, true, customFields.Fields["Boolean-field"], fmt.Sprintf("Boolean-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, 1234, customFields.Fields["Number-field"], fmt.Sprintf("Number-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, "foobar", customFields.Fields["String-field"], fmt.Sprintf("String-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, map[string]interface{}{"en": "Localized String", "fr": "Chaîne localisée"}, customFields.Fields["LocalizedString-field"], fmt.Sprintf("LocalizedString-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, "value2", customFields.Fields["Enum-field"], fmt.Sprintf("Enum-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, "value1", customFields.Fields["LocalizedEnum-field"], fmt.Sprintf("LocalizedEnum-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, map[string]interface{}{"centAmount": float64(150000), "currencyCode": "EUR", "fractionDigits": float64(2), "type": "centPrecision"}, customFields.Fields["Money-field"], fmt.Sprintf("Money-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, "2023-08-29", customFields.Fields["Date-field"], fmt.Sprintf("Date-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, "20:22:11.123", customFields.Fields["Time-field"], fmt.Sprintf("Time-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, "2023-08-29T20:22:11.123Z", customFields.Fields["DateTime-field"], fmt.Sprintf("DateTime-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, map[string]interface{}{"id": productType.ID, "typeId": "product-type"}, customFields.Fields["Reference-field"], fmt.Sprintf("Reference-field unexpected value for %s resource.", customFieldResourceType))
						assert.EqualValues(t, []any{"ENUM-1", "ENUM-3"}, customFields.Fields["Set-field"], fmt.Sprintf("Set-field unexpected value' for %s resource.", customFieldResourceType))
						return nil
					},
				),
			},
		}

		// Remove Custom fields from the resource one by one
		for index := range customFieldTypes {
			var customFieldTypesReduced = []string{}
			for i := range customFieldTypes {
				if i == index {
					continue
				}
				customFieldTypesReduced = append(customFieldTypesReduced, customFieldTypes[i])
			}
			fieldTypeValue := customFieldTypes[index]

			customFieldsAccTestSteps = append(customFieldsAccTestSteps, resource.TestStep{
				Config: getResourceConfig(customFieldResourceType, resourceShortName, resourceKey, customFieldTypesReduced),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						customFields, err := getResourceCustomFields(s, customFieldResourceType, resourceFullName)
						if err != nil {
							return err
						}
						assert.Nil(t, customFields.Fields[fmt.Sprintf("%s-field", fieldTypeValue)], fmt.Sprintf("%s-field expected to be removed.", fieldTypeValue))
						return nil
					},
				),
			})
		}

		// Remove all Custom fields from the resource
		customFieldsAccTestSteps = append(customFieldsAccTestSteps, resource.TestStep{
			Config: getResourceConfig(customFieldResourceType, resourceShortName, resourceKey, []string{}),
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					customFields, err := getResourceCustomFields(s, customFieldResourceType, resourceFullName)
					if err != nil {
						return err
					}
					assert.Nil(t, customFields, fmt.Sprintf("%v-field expected to be nil.", customFields))
					return nil
				},
			),
		})

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps:     customFieldsAccTestSteps,
		})
	}
}

func getResourceConfig(resourceType, resourceName, resourceKey string, customFields []string) string {
	// Load templates
	tpl, err := template.ParseGlob("testdata/custom_fields_test/*")
	if err != nil {
		panic(err)
	}

	templateData := map[string]any{
		"resource_type": resourceType,
		"resource_name": resourceName,
		"resource_key":  resourceKey,
		"custom":        customFields,
	}

	var out bytes.Buffer
	err = tpl.ExecuteTemplate(&out, "main", templateData)
	if err != nil {
		panic(err)
	}

	return out.String()
}

func getResourceCustomFields(s *terraform.State, resourceType, identifier string) (*platform.CustomFields, error) {
	switch resourceType {
	case "commercetools_channel":
		channel, err := testGetChannel(s, identifier)
		return channel.Custom, err
	case "commercetools_cart_discount":
		cartDiscount, err := testGetCartDiscount(s, identifier)
		return cartDiscount.Custom, err
	case "commercetools_category":
		category, err := testGetCategory(s, identifier)
		return category.Custom, err
	case "commercetools_customer_group":
		customerGroup, err := testGetCustomerGroup(s, identifier)
		return customerGroup.Custom, err
	case "commercetools_discount_code":
		discountCode, err := testGetDiscountCode(s, identifier)
		return discountCode.Custom, err
	case "commercetools_shipping_method":
		shippingMethod, err := testGetShippingMethod(s, identifier)
		return shippingMethod.Custom, err
	case "commercetools_store":
		store, err := testGetStore(s, identifier)
		return store.Custom, err
	default:
		panic(fmt.Sprintf("Unknown resource type %s", resourceType))
	}
}

func testGetCategory(s *terraform.State, identifier string) (*platform.Category, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Category %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.Categories().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func testGetShippingMethod(s *terraform.State, identifier string) (*platform.ShippingMethod, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Shipping Method %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.ShippingMethods().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
