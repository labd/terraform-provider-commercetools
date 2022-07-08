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

type TestTypeFieldData struct {
	Name string
	Type string
}

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

func TestExpandTypeFieldDefinitionItem(t *testing.T) {
	input := map[string]any{
		"name": "test",
		"label": map[string]any{
			"en": "Test",
			"nl": "Test",
		},
		"type": []any{
			map[string]any{
				"name": "String",
			},
		},
		"required":   false,
		"input_hint": "SingleLine",
	}

	_, err := expandTypeFieldDefinitionItem(input)
	if err != nil {
		t.Error("Got an unexpected error")
	}
}

func TestExpandTypeFieldType(t *testing.T) {
	// Test Boolean
	input := map[string]interface{}{
		"name": "Boolean",
	}
	result, err := expandTypeFieldType(input)
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
	_, err = expandTypeFieldType(input)
	if err == nil {
		t.Error("No error returned while Enum requires values")
	}
	inputValue := make([]interface{}, 2)
	inputValue[0] = map[string]interface{}{"key": "value1", "label": "Value 1"}
	inputValue[1] = map[string]interface{}{"key": "value2", "label": "Value 2"}
	input = map[string]interface{}{
		"name":  "Enum",
		"value": inputValue,
	}
	result, err = expandTypeFieldType(input)
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
	_, err = expandTypeFieldType(input)
	if err == nil {
		t.Error("No error returned while Reference requires reference_type_id")
	}
	input = map[string]interface{}{
		"name":              "Reference",
		"reference_type_id": "product",
	}
	result, err = expandTypeFieldType(input)
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
	_, err = expandTypeFieldType(input)
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
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}
						assert.EqualValues(t, resource.Key, key)
						return nil
					},
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
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(
						resourceName, "field.0.name", "skype_name"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.name", "existing_enum"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.type.0.element_type.0.value.#", "2"),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}
						assert.EqualValues(t, resource.Key, key)
						return nil
					},
				),
			},
			{
				Config: testAccTypeUpdateWithID(identifier, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "field.#", "12"),
					resource.TestCheckResourceAttr(
						resourceName, "field.3.name", "icq_uin"),
					resource.TestCheckResourceAttr(
						resourceName, "field.4.name", "testing"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.name", "existing_enum"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.type.0.element_type.0.value.#", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "field.1.type.0.element_type.0.value.1.label", "Evening Changed"),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}
						assert.EqualValues(t, resource.Key, key)
						return nil
					},
				),
			},
		},
	})
}

func TestAccTypes_FieldOrderUpdates(t *testing.T) {
	key := "acctest-type"
	identifier := "acctest_type"
	resourceName := fmt.Sprintf("commercetools_type.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigFields(
					key, "acctest_type",
					[]TestTypeFieldData{
						{Name: "field-one", Type: "String"},
						{Name: "field-two", Type: "String"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TypeTextInputHintSingleLine
						expected := []platform.FieldDefinition{
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-one",
								Label:     platform.LocalizedString{"en": "field-one"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-two",
								Label:     platform.LocalizedString{"en": "field-two"},
								InputHint: &SingleText,
							},
						}
						assert.EqualValues(t, resource.Key, key)
						assert.EqualValues(t, expected, resource.FieldDefinitions)
						return nil
					},
				),
			},
			{
				Config: testAccConfigFields(
					key, "acctest_type",
					[]TestTypeFieldData{
						{Name: "field-one", Type: "String"},
						{Name: "field-two", Type: "String"},
						{Name: "field-three", Type: "String"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TypeTextInputHintSingleLine
						expected := []platform.FieldDefinition{
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-one",
								Label:     platform.LocalizedString{"en": "field-one"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-two",
								Label:     platform.LocalizedString{"en": "field-two"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-three",
								Label:     platform.LocalizedString{"en": "field-three"},
								InputHint: &SingleText,
							},
						}
						assert.EqualValues(t, resource.Key, key)
						assert.EqualValues(t, expected, resource.FieldDefinitions)
						return nil
					},
				),
			},
			{
				Config: testAccConfigFields(
					key, "acctest_type",
					[]TestTypeFieldData{
						{Name: "field-one", Type: "String"},
						{Name: "field-three", Type: "String"},
						{Name: "field-two", Type: "String"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TypeTextInputHintSingleLine
						expected := []platform.FieldDefinition{
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-one",
								Label:     platform.LocalizedString{"en": "field-one"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-three",
								Label:     platform.LocalizedString{"en": "field-three"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-two",
								Label:     platform.LocalizedString{"en": "field-two"},
								InputHint: &SingleText,
							},
						}

						assert.EqualValues(t, resource.Key, key)
						assert.EqualValues(t, expected, resource.FieldDefinitions)
						return nil
					},
				),
			},
			{
				Config: testAccConfigFields(
					key, "acctest_type",
					[]TestTypeFieldData{
						{Name: "field-one", Type: "String"},
						{Name: "field-four", Type: "String"},
						{Name: "field-three", Type: "String"},
						{Name: "field-two", Type: "String"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TypeTextInputHintSingleLine
						expected := []platform.FieldDefinition{
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-one",
								Label:     platform.LocalizedString{"en": "field-one"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-four",
								Label:     platform.LocalizedString{"en": "field-four"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-three",
								Label:     platform.LocalizedString{"en": "field-three"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-two",
								Label:     platform.LocalizedString{"en": "field-two"},
								InputHint: &SingleText,
							},
						}

						assert.EqualValues(t, resource.Key, key)
						assert.EqualValues(t, expected, resource.FieldDefinitions)
						return nil
					},
				),
			},
			{
				Config: testAccConfigFields(
					key, "acctest_type",
					[]TestTypeFieldData{
						{Name: "field-one", Type: "String"},
						{Name: "field-two", Type: "String"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						resource, err := testGetType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TypeTextInputHintSingleLine
						expected := []platform.FieldDefinition{
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-one",
								Label:     platform.LocalizedString{"en": "field-one"},
								InputHint: &SingleText,
							},
							{
								Type:      platform.CustomFieldStringType{},
								Name:      "field-two",
								Label:     platform.LocalizedString{"en": "field-two"},
								InputHint: &SingleText,
							},
						}

						assert.EqualValues(t, resource.Key, key)
						assert.EqualValues(t, expected, resource.FieldDefinitions)
						return nil
					},
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
				value {
					key = "day"
					label = "Daytime"
				}
				value {
					key = "evening"
					label = "Evening"
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
			`,
				map[string]interface{}{
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
					value {
						key = "day"
						label = "Daytime"
					}
					value {
						key = "evening"
						label = "Evening"
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
						value {
							key = "day"
							label = "Daytime"
						}
						value {
							key = "evening"
							label = "Evening Changed"
						}
						value {
							key = "later"
							label = "later"
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

		}`,
		map[string]any{
			"identifier": identifier,
			"key":        key,
			"newFields":  newFieldsBuffer.String(),
		})
}

func testAccConfigFields(key, identifier string, fields []TestTypeFieldData) string {
	return hclTemplate(`
		resource "commercetools_type" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = { "en": "{{ .key }}" }
			resource_type_ids = ["customer"]

			{{range $t := .fields}}
			field {
				name = "{{ $t.Name }}"
				label = { en = "{{ $t.Name }}" }
				type { name = "{{ $t.Type }}" }
			}
			{{end}}
		}
		`, map[string]any{
		"key":        key,
		"identifier": identifier,
		"fields":     fields,
	},
	)

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
