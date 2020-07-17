package commercetools

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func TestFieldTypeElement(t *testing.T) {
	elem := fieldTypeElement(true)
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
		t.Error("element_type Sxhema should not include another 'element_type' field")
	}
}

func TestResourceTypeGetFieldDefinition(t *testing.T) {
	input := map[string]interface{}{
		"name": "test",
		"label": map[string]interface{}{
			"en": "Test",
			"nl": "Test",
		},
		"type": []interface{}{
			map[string]interface{}{
				"name": "String",
			},
		},
		"required":   false,
		"input_hint": "SingleLine",
	}

	_, err := resourceTypeGetFieldDefinition(input)
	if err != nil {
		t.Error("Got an unexpected error")
	}
}

func TestGetFieldType(t *testing.T) {
	// Test Boolean
	input := map[string]interface{}{
		"name": "Boolean",
	}
	result, err := getFieldType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := result.(commercetools.CustomFieldBooleanType); !ok {
		t.Error("Expected Boolean type")
	}

	// Test Enum
	input = map[string]interface{}{
		"name": "Enum",
	}
	_, err = getFieldType(input)
	if err == nil {
		t.Error("No error returned while Enum requires values")
	}
	input = map[string]interface{}{
		"name": "Enum",
		"values": map[string]interface{}{
			"value1": "Value 1",
			"value2": "Value 2",
		},
	}
	result, err = getFieldType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if field, ok := result.(commercetools.CustomFieldEnumType); ok {
		assert.ElementsMatch(t, field.Values, []commercetools.CustomFieldEnumValue{
			{Key: "value1", Label: "Value 1"},
			{Key: "value2", Label: "Value 2"},
		})
	} else {
		t.Error("Expected Enum type")
	}

	// Test Reference
	input = map[string]interface{}{
		"name": "Reference",
	}
	_, err = getFieldType(input)
	if err == nil {
		t.Error("No error returned while Reference requires reference_type_id")
	}
	input = map[string]interface{}{
		"name":              "Reference",
		"reference_type_id": "product",
	}
	result, err = getFieldType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if field, ok := result.(commercetools.CustomFieldReferenceType); ok {
		assert.EqualValues(t, field.ReferenceTypeID, "product")
	} else {
		t.Error("Expected Reference type")
	}

	// Test Set
	input = map[string]interface{}{
		"name": "Set",
	}
	_, err = getFieldType(input)
	if err == nil {
		t.Error("No error returned while Set requires element_type")
	}
}

func TestAccTypes_basic(t *testing.T) {
	name := "acctest_type"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccTypeExists("acctest_type"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "key", name),
				),
			},
		},
	})
}

func TestAccTypes_UpdateWithID(t *testing.T) {
	name := "acctest_type"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccTypeExists("acctest_type"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "key", name),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "field.1.name", "existing_enum"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "field.1.type.0.element_type.0.values.%", "2"),
				),
			},
			{
				Config: testAccTypeUpdateWithID(name),
				Check: resource.ComposeTestCheckFunc(
					testAccTypeExists("acctest_type"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "key", name),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "field.#", "11"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "field.2.name", "existing_enum"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "field.2.type.0.element_type.0.values.%", "3"),
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "field.2.type.0.element_type.0.values.evening", "Evening Changed"),
				),
			},
		},
	})
}

func testAccTypeConfig(name string) string {
	return fmt.Sprintf(`
resource "commercetools_type" "%s" {
	key = "%s"
	name = {
		en = "Contact info"
		nl = "Contact informatie"
	}
	description = {
		en = "All things related communication"
		nl = "Alle communicatie-gerelateerde zaken"
	}

	resource_type_ids = ["customer"]

	field {
		name = "skype_name"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "String"
		}
	}

	field {
		name = "existing_enum"
		label = {
			en = "existing enum"
			de = "existierendes enum"
		}
		type {
			name = "Set" 
			element_type {
				name = "Enum"
				values = {
					day = "Daytime"
					evening = "Evening"
				}
			}
		}
	}

}`, name, name)
}

func testAccTypeUpdateWithID(name string) string {
	newFields := []string{
		"Boolean",
		"LocalizedString",
		"Number",
		"Money",
		"Date",
		"Time",
		"DateTime",
	}
	var newFieldsBuffer bytes.Buffer
	for _, newType := range newFields {
		newFieldsBuffer.WriteString(
			fmt.Sprintf(`
		field {
			name = "%[1]s"
			label = {
				en = "%[1]s"
				nl = "%[1]s"
			}

			type {
				name = "%[1]s"
			}
		}
		`, newType))
	}

	return fmt.Sprintf(`
resource "commercetools_type" "%s" {
	key = "%s"
	name = {
		en = "Contact info"
		nl = "Contact informatie"
	}
	description = {
		en = "All things related communication"
		nl = "Alle communicatie-gerelateerde zaken"
	}

	resource_type_ids = ["customer"]

	field {
		name = "skype_name"
		label = {
			en = "Skype name"
			nl = "Skype naam"
		}
		type {
			name = "String"
		}
	}

	field {
		name = "new_enum"
		label = {
			en = "new enum"
			nl = "nieuwe enum"
		}
		type {
			name = "Enum"
			values = {
				day = "Daytime"
				evening = "Evening"
			}
		}
	}

	field {
		name = "existing_enum"
		label = {
			en = "existing enum"
			de = "existierendes enum"
		}
		type {
			name = "Set" 
			element_type {
				name = "Enum"
				values = {
					day = "Daytime"
					evening = "Evening Changed"
					later   = "later"
				}
			}
		}
	}

	field {
		name = "new_localized_enum"
		input_hint = "MultiLine"
		label = {
			en = "New localized enum"
			nl = "Nieuwe localized enum"
		}
		type {
			name = "LocalizedEnum"
			localized_value {
				key = "phone"
				label = {
					en = "Phone"
					nl = "Telefoon"
				}
			}
			localized_value {
				key = "skype"
				label = {
					en = "Skype"
					nl = "Skype"
				}
			}
		}
	}

	%s
}`, name, name, newFieldsBuffer.String())
}

func testAccTypeExists(n string) resource.TestCheckFunc {
	name := fmt.Sprintf("commercetools_type.%s", n)
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Type ID is set")
		}

		client := getClient(testAccProvider.Meta())
		result, err := client.TypeGetWithID(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("Type not found")
		}

		return nil
	}
}

func testAccCheckTypesDestroy(s *terraform.State) error {
	// TODO: Implement
	return nil
}
