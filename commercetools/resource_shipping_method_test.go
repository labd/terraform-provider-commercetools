package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccShippingMethod_createAndUpdateWithID(t *testing.T) {

	name := "test method"
	key := "test-method"
	description := "test shipping method description"
	predicate := "1 = 1"

	newName := "new test method"
	newKey := "new-test-method"
	newDescription := "new test shipping method description"
	newPredicate := "2 = 2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodConfig(name, key, description, false, false, predicate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "key", key,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "description", description,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "is_default", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "predicate", predicate,
					),
				),
			},
			{
				Config: testAccShippingMethodConfig(newName, newKey, newDescription, true, true, newPredicate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "name", newName,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "key", newKey,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "description", newDescription,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "is_default", "true",
					),
					resource.TestCheckResourceAttrSet(
						"commercetools_shipping_method.standard", "tax_category_id",
					),
					resource.TestCheckResourceAttrSet(
						"commercetools_shipping_method.standard", "predicate", newPredicate,
					),
				),
			},
		},
	})
}

func testAccShippingMethodConfig(name string, key string, description string, isDefault bool, setTaxCategory bool, predicate string) string {
	taxCategoryReference := ""
	if setTaxCategory {
		taxCategoryReference = "tax_category_id = \"${commercetools_tax_category.test.id}\""
	}
	return fmt.Sprintf(`
resource "commercetools_tax_category" "test" {
	name = "test"
	key = "test"
	description = "test"
	predicate = "1=1"
}

resource "commercetools_shipping_method" "standard" {
	name = "%s"
	key = "%s"
	description = "%s"
	is_default = "%t"
	predicate = "%s"
    `, name, key, description, isDefault, predicate) + taxCategoryReference + "\n}\n"
}

func testAccCheckShippingMethodDestroy(s *terraform.State) error {
	return nil
}
