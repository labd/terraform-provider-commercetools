package commercetools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
			{Key: "value1", Label: "Value 1"},
			{Key: "value2", Label: "Value 2"},
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
						"commercetools_product_type.acctest_product_type", "key", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "name", "Shipping info",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "description", "All things related shipping",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.0.name", "location",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.0.label.en", "Location",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.0.label.nl", "Locatie",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.0.type.0.name", "text",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.1.type.0.localized_value.0.label.en", "Snack",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.1.type.0.localized_value.0.label.nl", "maaltijd",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.2.type.0.element_type.0.localized_value.0.label.en", "Breakfast",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.2.type.0.element_type.0.localized_value.1.label.en", "Lunch",
					),
				),
			},
			{
				Config: testAccProductTypeConfigLabelChange(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "key", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "name", "Shipping info",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "description", "All things related shipping",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.0.name", "location",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.0.label.en", "Location change",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.1.type.0.localized_value.0.label.en", "snack",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.1.type.0.localized_value.0.label.nl", "nomnom",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.1.type.0.localized_value.0.label.de", "happen",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.2.type.0.element_type.0.localized_value.0.label.en", "Breakfast",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.2.type.0.element_type.0.localized_value.1.label.en", "Lunch",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.2.type.0.element_type.0.localized_value.0.label.de", "Frühstück",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.2.type.0.element_type.0.localized_value.1.label.de", "Mittagessen",
					),
				),
			},
		},
	})
}

func testAccProductTypeConfigLabelChange(name string) string {
	return fmt.Sprintf(`
resource "commercetools_product_type" "acctest_product_type" {
	key = "%s"
	name = "Shipping info"
	description = "All things related shipping"

	attribute {
		name = "location"
		label = {
			en = "Location change"
			nl = "Locatie"
		}
		type {
			name = "text"
		}
	}

	attribute {
		name = "meal"
		label = {
			en = "meal"
			nl = "maaltijd"
		}

		type {
			name = "lenum"

            localized_value {
			  key = "snack"

			  label = {
				en = "snack"
				nl = "nomnom"
				de = "happen"
			  }
			}
		}
	}

	attribute {
		name = "types"
		label = {
			en = "meal types"
		}

		type {
			name = "set"
			element_type {
				name = "lenum"

				localized_value {
				  key = "breakfast"
	
				  label = {
					en = "Breakfast"
					de = "Frühstück"
				  }
				}

				localized_value {
				  key = "lunch"
	
				  label = {
					en = "Lunch"
					de = "Mittagessen"
				  }
				}
			}
		}
	}

}`, name)
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

	attribute {
		name = "meal"
		label = {
			en = "meal"
			nl = "maaltijd"
		}

		type {
			name = "lenum"

			localized_value {
			  key = "snack"

			  label = {
				en = "Snack"
				nl = "maaltijd"
			  }
			}
		}
	}

	attribute {
		name = "types"
		label = {
			en = "meal types"
		}

		type {
			name = "set"
			element_type {
				name = "lenum"

				localized_value {
				  key = "breakfast"
	
				  label = {
					en = "Breakfast"
				  }
				}

				localized_value {
				  key = "lunch"
	
				  label = {
					en = "Lunch"
				  }
				}
			}
		}
	}
}`, name)
}

func testAccCheckProductTypesDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*commercetools.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product_type" {
			continue
		}
		response, err := conn.ProductTypeGetWithID(context.Background(), rs.Primary.ID)
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product type (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
		if !strings.Contains(err.Error(), "was not found") {
			return err
		}
	}
	return nil
}
