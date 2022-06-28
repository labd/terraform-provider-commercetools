package commercetools

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
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
	if _, ok := result.(platform.CustomFieldBooleanType); !ok {
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
	if field, ok := result.(platform.CustomFieldEnumType); ok {
		assert.ElementsMatch(t, field.Values, []platform.CustomFieldEnumValue{
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
	if field, ok := result.(platform.CustomFieldReferenceType); ok {
		assert.EqualValues(t, field.ReferenceTypeId, "product")
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
	key := "acctest-type"
	identifier := "acctest_type"
	resourceName := fmt.Sprintf("commercetools_type.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig(identifier, key),
				Check: resource.ComposeTestCheckFunc(
					testAccTypeExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "key", key),
				),
			},
		},
	})
}

func TestAccTypes_UpdateWithID(t *testing.T) {
	key := "acctest-type"
	identifier := "acctest_type"
	resourceName := fmt.Sprintf("commercetools_type.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig(identifier, key),
				Check: resource.ComposeTestCheckFunc(
					testAccTypeExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "key", key),
					resource.TestCheckResourceAttr(
						resourceName, "field.0.name", "skype_name"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.name", "existing_enum"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.type.0.element_type.0.values.%", "2"),
				),
			},
			{
				Config: testAccTypeUpdateWithID(identifier, key),
				Check: resource.ComposeTestCheckFunc(
					testAccTypeExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "key", key),
					resource.TestCheckResourceAttr(
						resourceName, "field.#", "12"),
					resource.TestCheckResourceAttr(
						resourceName, "field.3.name", "icq_uin"),
					resource.TestCheckResourceAttr(
						resourceName, "field.4.name", "testing"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.name", "existing_enum"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.type.0.element_type.0.values.%", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.type.0.element_type.0.values.evening", "Evening Changed"),
				),
			},
		},
	})
}

func testAccTypeConfig(identifier, key string) string {
	return hclTemplate(`
resource "commercetools_type" "{{ .identifier }}" {
	key = "{{ .key }}"
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

}`, map[string]interface{}{
		"identifier": identifier,
		"key":        key,
	})
}

func testAccTypeUpdateWithID(identifier, key string) string {
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
			hclTemplate(`
		field {
			name = "{{ .name }}"
			label = {
				en = "{{ .label }}"
				nl = "{{ .label }}"
			}

			type {
				name = "{{ .typeName }}"
			}
		}
		`, map[string]interface{}{
				"name":     newType,
				"label":    newType,
				"typeName": newType,
			}))
	}

	return hclTemplate(`
resource "commercetools_type" "{{ .identifier }}" {
	key = "{{ .key }}"
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

	field {
		name = "icq_uin"
		label = {
			en = "UIN"
		}
		type {
			name = "String"
		}
	}

	field {
		name = "testing"
		label = {
			en = "test"
		}
		type {
			name = "String"
		}
	}

	{{ .newFields }}

}`, map[string]interface{}{
		"identifier": identifier,
		"key":        key,
		"newFields":  newFieldsBuffer.String(),
	})
}

func testAccTypeExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetType(s, fmt.Sprintf("commercetools_type.%s", n))
		return err
	}
}

func testAccCheckTypesDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_type" {
			continue
		}
		response, err := client.Types().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("type (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testGetType(s *terraform.State, identifier string) (*platform.Type, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Type %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.Types().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
