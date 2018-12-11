package commercetools

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func TestAttributeTypeElement(t *testing.T) {
	elem := attributeTypeElement(true)
	elemType, ok := elem.Schema["element_type"]

	if !ok {
		t.Error("element_type does not appear in the Schema")
	}

	elemTypeResource := elemType.Elem.(*schema.Resource)

	// The element_type itself may not contain an 'element_type'.
	// This is because we don't allow infinite nested 'Set' elements
	if _, ok := elemTypeResource.Schema["name"]; !ok {
		t.Error("element_type Schema does not contain 'name' field")
	}
	if _, ok := elemTypeResource.Schema["element_type"]; ok {
		t.Error("element_type Schema should not include another 'element_type' field")
	}
}

func TestGetAttributeType(t *testing.T) {
	// Test Boolean
	input := map[string]interface{}{
		"name": "boolean",
	}
	result, err := getAttributeType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := result.(commercetools.AttributeBooleanType); !ok {
		t.Error("Expected Boolean type")
	}

	// Test Enum
	input = map[string]interface{}{
		"name": "enum",
	}
	_, err = getAttributeType(input)
	if err == nil {
		t.Error("No error returned while enum requires values")
	}
	input = map[string]interface{}{
		"name": "enum",
		"values": map[string]interface{}{
			"value1": "Value 1",
			"value2": "Value 2",
		},
	}
	result, err = getAttributeType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if field, ok := result.(commercetools.AttributeEnumType); ok {
		assert.ElementsMatch(t, field.Values, []commercetools.AttributePlainEnumValue{
			commercetools.AttributePlainEnumValue{Key: "value1", Label: "Value 1"},
			commercetools.AttributePlainEnumValue{Key: "value2", Label: "Value 2"},
		})
	} else {
		t.Error("Expected Enum type")
	}

	// Test Reference
	input = map[string]interface{}{
		"name": "reference",
	}
	_, err = getAttributeType(input)
	if err == nil {
		t.Error("No error returned while Reference requires reference_type_id")
	}
	input = map[string]interface{}{
		"name":              "reference",
		"reference_type_id": "product",
	}
	result, err = getAttributeType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if field, ok := result.(commercetools.AttributeReferenceType); ok {
		assert.EqualValues(t, field.ReferenceTypeID, "product")
	} else {
		t.Error("Expected Reference type")
	}

	// Test Set
	input = map[string]interface{}{
		"name": "set",
	}
	_, err = getAttributeType(input)
	if err == nil {
		t.Error("No error returned while set requires element_type")
	}
}

func TestAccProductTypes_basic(t *testing.T) {
	name := "acctest_producttype"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProductTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductTypeConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "key", name),
				),
			},
		},
	})
}

func testAccProductTypeConfig(name string) string {
	return fmt.Sprintf(`
resource "commercetools_product_type" "acctest_product_type" {
	key = "%s"
	name = "Shipping info"
	description = "All things related shipping"

	attribute {
		name = "location"
		label = {
			en = "Location"
			nl = "Locatie"
		}
		type {
			name = "text"
		}
	}
}`, name)
}

func testAccCheckProductTypesDestroy(s *terraform.State) error {
	// TODO: Implement
	return nil
}
