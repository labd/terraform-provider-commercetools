package commercetools

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

type TestProductTypeAttrData struct {
	Name        string
	Type        string
	Level       string
	Values      []TestProductTypeEnumValue
	ElementType *TestProductTypeElementType
}

type TestProductTypeEnumValue struct {
	Key   string
	Label string
}

type TestProductTypeElementType struct {
	Name   string
	Values []TestProductTypeEnumValue
}

func TestResourceProductTypeValidateAttribute(t *testing.T) {
	o := []any{
		map[string]any{
			"name": "attr-one",
			"type": []any{
				map[string]any{
					"name": "text",
				},
			},
		},
	}
	n := []any{
		map[string]any{
			"name": "attr-one",
			"type": []any{
				map[string]any{
					"name": "Boolean",
				},
			},
		},
	}
	err := resourceProductTypeValidateAttribute(o, n)
	assert.NotNil(t, err)
}

func TestResourceProductTypeValidateAttributeSet(t *testing.T) {
	o := []any{
		map[string]any{
			"name": "attr-one",
			"type": []any{
				map[string]any{
					"name": "Set",
					"element_type": []any{
						map[string]any{
							"name": "text",
						},
					},
				},
			},
		},
	}
	n := []any{
		map[string]any{
			"name": "attr-one",
			"type": []any{
				map[string]any{
					"name": "Set",
					"element_type": []any{
						map[string]any{
							"name": "Enum",
						},
					},
				},
			},
		},
	}
	err := resourceProductTypeValidateAttribute(o, n)
	assert.NotNil(t, err)
}

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
		t.Error("element_type Schema does not contain 'name' attribute")
	}
	if _, ok := elemTypeResource.Schema["element_type"]; ok {
		t.Error("element_type Schema should not include another 'element_type' attribute")
	}
}

func TestExpandProductTypeAttributeType(t *testing.T) {
	// Test Boolean
	input := map[string]any{
		"name": "boolean",
	}
	result, err := expandProductTypeAttributeType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := result.(platform.AttributeBooleanType); !ok {
		t.Error("Expected Boolean type")
	}

	// Test Enum
	input = map[string]any{
		"name": "enum",
	}
	_, err = expandProductTypeAttributeType(input)
	if err == nil {
		t.Error("No error returned while enum requires values")
	}
	inputValue := make([]any, 2)
	inputValue[0] = map[string]any{"key": "value1", "label": "Value 1"}
	inputValue[1] = map[string]any{"key": "value2", "label": "Value 2"}
	input = map[string]any{
		"name":  "enum",
		"value": inputValue,
	}
	result, err = expandProductTypeAttributeType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if attr, ok := result.(platform.AttributeEnumType); ok {
		assert.ElementsMatch(t, attr.Values, []platform.AttributePlainEnumValue{
			{Key: "value1", Label: "Value 1"},
			{Key: "value2", Label: "Value 2"},
		})
	} else {
		t.Error("Expected Enum type")
	}

	// Test Reference
	input = map[string]any{
		"name": "reference",
	}
	_, err = expandProductTypeAttributeType(input)
	if err == nil {
		t.Error("No error returned while Reference requires reference_type_id")
	}
	input = map[string]any{
		"name":              "reference",
		"reference_type_id": "product",
	}
	result, err = expandProductTypeAttributeType(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if attr, ok := result.(platform.AttributeReferenceType); ok {
		assert.EqualValues(t, attr.ReferenceTypeId, "product")
	} else {
		t.Error("Expected Reference type")
	}

	// Test Set
	input = map[string]any{
		"name": "set",
	}
	_, err = expandProductTypeAttributeType(input)
	if err == nil {
		t.Error("No error returned while set requires element_type")
	}
}

func TestAccProductTypes_basic(t *testing.T) {
	key := "acctest-producttype"
	identifier := "acctest_producttype"
	resourceName := "commercetools_product_type.acctest_producttype"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckProductTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductTypeConfig(identifier, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}
						assert.EqualValues(t, *r.Key, key)
						return nil
					},
					resource.TestCheckResourceAttr(
						resourceName, "name", "Shipping info",
					),
					resource.TestCheckResourceAttr(
						resourceName, "description", "All things related shipping",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.#", "3",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.0.name", "location",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.0.label.en", "Location",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.0.label.nl", "Locatie",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.0.type.0.name", "text",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.1.type.0.localized_value.0.label.en", "Snack",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.1.type.0.localized_value.0.label.nl", "maaltijd",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.2.type.0.element_type.0.localized_value.0.label.en", "Breakfast",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.2.type.0.element_type.0.localized_value.1.label.en", "Lunch",
					),
				),
			},
			{
				Config: testAccProductTypeConfigLabelChange(identifier, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName, "key", key,
					),
					resource.TestCheckResourceAttr(
						resourceName, "name", "Shipping info",
					),
					resource.TestCheckResourceAttr(
						resourceName, "description", "All things related shipping",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.#", "3",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.0.name", "location",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.0.label.en", "Location change",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.1.type.0.localized_value.0.label.en", "snack",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.1.type.0.localized_value.0.label.nl", "nomnom",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.1.type.0.localized_value.0.label.de", "happen",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.2.type.0.element_type.0.localized_value.0.label.en", "Breakfast",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.2.type.0.element_type.0.localized_value.1.label.en", "Lunch",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.2.type.0.element_type.0.localized_value.0.label.de", "Frühstück",
					),
					resource.TestCheckResourceAttr(
						resourceName, "attribute.2.type.0.element_type.0.localized_value.1.label.de", "Mittagessen",
					),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}
						assert.EqualValues(t, *r.Key, key)
						return nil
					},
				),
			},
		},
	})
}

func TestAccProductTypes_AttributeOrderUpdates(t *testing.T) {
	key := "acctest-producttype"
	identifier := "acctest_producttype"
	resourceName := fmt.Sprintf("commercetools_product_type.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{Name: "attr-one", Type: "text"},
						{Name: "attr-two", Type: "text"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
						}
						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{Name: "attr-one", Type: "text"},
						{Name: "attr-two", Type: "text"},
						{Name: "attr-three", Type: "text"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-three",
								Label:               platform.LocalizedString{"en": "attr-three"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
						}
						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{Name: "attr-one", Type: "text"},
						{Name: "attr-three", Type: "text"},
						{Name: "attr-two", Type: "text"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-three",
								Label:               platform.LocalizedString{"en": "attr-three"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
						}

						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{Name: "attr-one", Type: "text"},
						{Name: "attr-four", Type: "text"},
						{Name: "attr-three", Type: "text"},
						{Name: "attr-two", Type: "text"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-four",
								Label:               platform.LocalizedString{"en": "attr-four"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-three",
								Label:               platform.LocalizedString{"en": "attr-three"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
						}

						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{Name: "attr-one", Type: "text"},
						{Name: "attr-two", Type: "text"},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
							{
								Type:                platform.AttributeTextType{},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
							},
						}

						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
		},
	})
}

func TestAccProductTypes_EnumValues(t *testing.T) {
	key := "acctest-producttype"
	identifier := "acctest_producttype"
	resourceName := fmt.Sprintf("commercetools_product_type.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{
							Name: "attr-one",
							Type: "enum",
							Values: []TestProductTypeEnumValue{
								{
									Key:   "value_1",
									Label: "Value-1",
								},
								{
									Key:   "value_2",
									Label: "Value-2",
								},
							}},
						{
							Name: "attr-two",
							Type: "set",
							ElementType: &TestProductTypeElementType{
								Name: "enum",
								Values: []TestProductTypeEnumValue{
									{
										Key:   "set_value_1",
										Label: "Set-value-1",
									},
									{
										Key:   "set_value_2",
										Label: "Set-value-2",
									},
								},
							},
						},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type: platform.AttributeEnumType{
									Values: []platform.AttributePlainEnumValue{
										{
											Key:   "value_1",
											Label: "Value-1",
										},
										{
											Key:   "value_2",
											Label: "Value-2",
										},
									},
								},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
							{
								Type: platform.AttributeSetType{
									ElementType: platform.AttributeEnumType{
										Values: []platform.AttributePlainEnumValue{
											{
												Key:   "set_value_1",
												Label: "Set-value-1",
											},
											{
												Key:   "set_value_2",
												Label: "Set-value-2",
											},
										},
									},
								},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
						}
						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{
							Name: "attr-one",
							Type: "enum",
							Values: []TestProductTypeEnumValue{
								{
									Key:   "value_2",
									Label: "Value-2",
								},
							},
						},
						{
							Name: "attr-two",
							Type: "set",
							ElementType: &TestProductTypeElementType{
								Name: "enum",
								Values: []TestProductTypeEnumValue{
									{
										Key:   "set_value_2",
										Label: "Set-value-2",
									},
								},
							},
						},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type: platform.AttributeEnumType{
									Values: []platform.AttributePlainEnumValue{
										{
											Key:   "value_2",
											Label: "Value-2",
										},
									},
								},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
							{
								Type: platform.AttributeSetType{
									ElementType: platform.AttributeEnumType{
										Values: []platform.AttributePlainEnumValue{
											{
												Key:   "set_value_2",
												Label: "Set-value-2",
											},
										},
									},
								},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
						}
						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
		},
	})
}

func TestAccProductTypes_sliced(t *testing.T) {
	t.Skip("Skipping test for large number of attributes")

	key := "acctest-producttype"
	identifier := "acctest_producttype"
	resourceName := "commercetools_product_type.acctest_producttype"

	var attributes []TestProductTypeAttrData
	for i := 0; i < 1000; i++ {
		attributes = append(attributes, TestProductTypeAttrData{
			Name: fmt.Sprintf("%d", i),
			Type: "text",
		})
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckProductTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAttributes(key, identifier, []TestProductTypeAttrData{}),
				Check: func(s *terraform.State) error {
					r, err := testGetProductType(s, resourceName)
					if err != nil {
						return err
					}
					assert.EqualValues(t, *r.Key, key)
					return nil
				},
			},
			{
				Config: testAccConfigAttributes(key, identifier, attributes),
				Check: func(s *terraform.State) error {
					r, err := testGetProductType(s, resourceName)
					if err != nil {
						return err
					}
					assert.EqualValues(t, *r.Key, key)
					return nil
				},
			},
		},
	})
}

func TestAccProductTypes_ProductVariant(t *testing.T) {
	key := "acctest-producttype"
	identifier := "acctest_producttype"
	resourceName := fmt.Sprintf("commercetools_product_type.%s", identifier)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTypesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{
							Name:  "attr-one",
							Type:  "enum",
							Level: string(platform.AttributeLevelEnumProduct),
							Values: []TestProductTypeEnumValue{
								{
									Key:   "value_1",
									Label: "Value-1",
								},
								{
									Key:   "value_2",
									Label: "Value-2",
								},
							}},
						{
							Name: "attr-two",
							Type: "set",
							ElementType: &TestProductTypeElementType{
								Name: "enum",
								Values: []TestProductTypeEnumValue{
									{
										Key:   "set_value_1",
										Label: "Set-value-1",
									},
									{
										Key:   "set_value_2",
										Label: "Set-value-2",
									},
								},
							},
						},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						r, err := testGetProductType(s, resourceName)
						if err != nil {
							return err
						}

						SingleText := platform.TextInputHintSingleLine
						expected := []platform.AttributeDefinition{
							{
								Type: platform.AttributeEnumType{
									Values: []platform.AttributePlainEnumValue{
										{
											Key:   "value_1",
											Label: "Value-1",
										},
										{
											Key:   "value_2",
											Label: "Value-2",
										},
									},
								},
								Name:                "attr-one",
								Label:               platform.LocalizedString{"en": "attr-one"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumProduct,
								InputTip:            nil,
							},
							{
								Type: platform.AttributeSetType{
									ElementType: platform.AttributeEnumType{
										Values: []platform.AttributePlainEnumValue{
											{
												Key:   "set_value_1",
												Label: "Set-value-1",
											},
											{
												Key:   "set_value_2",
												Label: "Set-value-2",
											},
										},
									},
								},
								Name:                "attr-two",
								Label:               platform.LocalizedString{"en": "attr-two"},
								InputHint:           SingleText,
								AttributeConstraint: platform.AttributeConstraintEnumNone,
								Level:               platform.AttributeLevelEnumVariant,
								InputTip:            nil,
							},
						}
						assert.EqualValues(t, *r.Key, key)
						assert.EqualValues(t, expected, r.Attributes)
						return nil
					},
				),
			},
			{
				Config: testAccConfigAttributes(
					key, "acctest_producttype",
					[]TestProductTypeAttrData{
						{
							Name:  "attr-one",
							Type:  "enum",
							Level: string(platform.AttributeLevelEnumProduct),
							Values: []TestProductTypeEnumValue{
								{
									Key:   "value_1",
									Label: "Value-1",
								},
								{
									Key:   "value_2",
									Label: "Value-2",
								},
							},
						},
						{
							Name:  "attr-two",
							Type:  "set",
							Level: string(platform.AttributeLevelEnumProduct),
							ElementType: &TestProductTypeElementType{
								Name: "enum",
								Values: []TestProductTypeEnumValue{
									{
										Key:   "set_value_1",
										Label: "Set-value-1",
									},
									{
										Key:   "set_value_2",
										Label: "Set-value-2",
									},
								},
							},
						},
					}),
				ExpectError: regexp.MustCompile("changing the level of an attribute is not supported in commercetools"),
			},
		},
	})
}

func testAccProductTypeConfigLabelChange(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_product_type" "{{ .identifier }}" {
			key = "{{ .key }}"
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
		}`, map[string]any{"key": key, "identifier": identifier})
}

func testAccProductTypeConfig(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_product_type" "{{ .identifier }}" {
			key = "{{ .key }}"
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
		}`, map[string]any{"key": key, "identifier": identifier})
}

func testAccConfigAttributes(key, identifier string, attrs []TestProductTypeAttrData) string {
	output := hclTemplate(`
		resource "commercetools_product_type" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = "{{ .key }}"
	
			{{range $t := .attributes}}
			attribute {
				name = "{{ $t.Name }}"
				label = { en = "{{ $t.Name }}" }
				{{ if ne $t.Level "" }}
				level = "{{ $t.Level }}"
				{{ end }}
				type {
					name = "{{ $t.Type }}"

					{{ if eq $t.Type "set" }}
					element_type {
						name = "{{ $t.ElementType.Name }}"
						{{ if $t.ElementType.Values }}
							{{ range $v := $t.ElementType.Values }}
								value {
									key = "{{ $v.Key }}"
									label = "{{ $v.Label }}"
								}
							{{ end }}
						{{ end }}
					}
					{{ end }}

					{{ if eq $t.Type "enum" }}
						{{ range $v := $t.Values }}
							value {
								key = "{{ $v.Key }}"
								label = "{{ $v.Label }}"
							}
						{{ end }}
					{{ end }}
				}

			}
			{{end}}
		}
		`, map[string]any{
		"key":        key,
		"identifier": identifier,
		"attributes": attrs,
	},
	)
	return output
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

func testGetProductType(s *terraform.State, identifier string) (*platform.ProductType, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("ProductType %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.ProductTypes().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
