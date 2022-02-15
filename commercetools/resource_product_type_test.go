package commercetools

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	if _, ok := result.(platform.AttributeBooleanType); !ok {
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
	if field, ok := result.(platform.AttributeEnumType); ok {
		assert.ElementsMatch(t, field.Values, []platform.AttributePlainEnumValue{
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
	if field, ok := result.(platform.AttributeReferenceType); ok {
		assert.EqualValues(t, field.ReferenceTypeId, "product")
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

	if os.Getenv("CTP_CLIENT_ID") == "unittest" {
		t.Skip("Skipping testing with mock server as the implementation can not handle order of localized enums")
	}

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
						"commercetools_product_type.acctest_product_type", "attribute.#", "6",
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
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.3.type.0.name", "lenum",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.3.type.0.localized_value.0.key", "cm",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.3.type.0.localized_value.1.key", "ml",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.4.type.0.name", "set",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.4.type.0.element_type.0.values.%", "5",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.5.type.0.name", "enum",
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
						"commercetools_product_type.acctest_product_type", "attribute.#", "6",
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

					func(s *terraform.State) error {
						if os.Getenv("CTP_CLIENT_ID") == "unittest" {
							t.Log("Skipping check of order as the mock server does not support this correctly")
							return nil
						}

						return resource.TestCheckResourceAttr(
							"commercetools_product_type.acctest_product_type", "attribute.3.type.0.localized_value.0.key", "ml",
						)(s)
					},
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.3.type.0.localized_value.0.key", "ml",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.3.type.0.localized_value.1.key", "cm",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.4.type.0.element_type.0.values.%", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_type.acctest_product_type", "attribute.5.type.0.name", "enum",
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

	attribute {
		label      = {
			"de-DE" = "Maßeinheit"
			"en"    = "Unit"
		}
		name       = "unit"
		type {
			name = "lenum"
            
			localized_value {
			  key = "ml"

			  label = {
				en = "ml"
				nl = "ml"
			  }
			}

			localized_value {
			  key = "cm"

			  label = {
				en = "cm"
				nl = "cm"
			  }
			}
		}
	}

	attribute {
		label      = {
			"de-DE" = "stores"
			"en"    = "stores"
		}
		name       = "onSale"
		type {
			name   = "set"
		   	element_type {
				name   = "enum"
				values = {
					"de"		 = "de"
					"not_de"     = "not_de"
				}
			}
		}
	}

	attribute {
		label      = {
			"de-DE" = "storesOrder"
			"en"    = "storesOrder"
		}
		name       = "storeOrder"
		type {
			name   = "enum"
			values = {
				"at" = "at"
				"de" = "de"
				"pl" = "pl"
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

	attribute {
		label      = {
			"de-DE" = "Maßeinheit"
			"en"    = "Unit"
		}
		name       = "unit"
		type {
			name = "lenum"
            
			localized_value {
			  key = "cm"

			  label = {
				en = "cm"
				nl = "cm"
			  }
			}

			localized_value {
			  key = "ml"

			  label = {
				en = "ml"
				nl = "ml"
			  }
			}
		}
	}

	attribute {
		label      = {
			"de-DE" = "stores"
			"en"    = "stores"
		}
		name       = "onSale"
		type {
			name   = "set"
		   	element_type {
				name   = "enum"
				values = {
					"AT"         = "AT"
					"DE"         = "DE"
					"PL"         = "PL"
					"de"		 = "de"
					"not_de"     = "not_de"
				}
			}
		}
	}

	attribute {
		label      = {
			"de-DE" = "storesOrder"
			"en"    = "storesOrder"
		}
		name       = "storeOrder"
		type {
			name   = "enum"
			values = {
				"pl" = "pl"
				"de" = "de"
				"at" = "at"
			}
		}
	}
}`, name)
}

func testAccCheckProductTypesDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product_type" {
			continue
		}
		response, err := client.ProductTypes().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product type (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
