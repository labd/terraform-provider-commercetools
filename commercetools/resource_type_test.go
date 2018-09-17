package commercetools

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/types"
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

func TestResourceTypeCreateFieldLookup(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"name":  "name1",
			"value": "Value 1",
		},
		map[string]interface{}{
			"name":  "name2",
			"value": "Value 2",
		},
	}
	result := resourceTypeCreateFieldLookup(input)
	if _, ok := result["name1"]; !ok {
		t.Error("Could not lookup name1")
	}
	if _, ok := result["name2"]; !ok {
		t.Error("Could not lookup name1")
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
	if _, ok := result.(types.BooleanType); !ok {
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
	if field, ok := result.(types.EnumType); ok {
		assert.EqualValues(t, field.Values, []commercetools.EnumValue{
			commercetools.EnumValue{Key: "value1", Label: "Value 1"},
			commercetools.EnumValue{Key: "value2", Label: "Value 2"},
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
	if field, ok := result.(types.ReferenceType); ok {
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
					resource.TestCheckResourceAttr(
						"commercetools_type.acctest_type", "key", name),
				),
			},
		},
	})
}

func testAccTypeConfig(name string) string {
	return fmt.Sprintf(`
resource "commercetools_type" "acctest_type" {
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
}`, name)
}

func testAccCheckTypesDestroy(s *terraform.State) error {
	// TODO: Implement
	return nil
}
