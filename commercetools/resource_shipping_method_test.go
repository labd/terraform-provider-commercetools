package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccShippingMethod_createAndUpdateWithID(t *testing.T) {

	name := "test sh method"
	key := "test-sh-method"
	description := "test shipping method description"
	predicate := "1 = 1"

	newName := "new test sh method"
	newKey := "new-test-sh-method"
	newDescription := "new test shipping method description"
	newPredicate := "2 = 2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodConfig(name, key, description, description, false, true, predicate),
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
						"commercetools_shipping_method.standard", "localized_description.en", description,
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
				Config: testAccShippingMethodConfig(newName, newKey, newDescription, newDescription, true, true, newPredicate),
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
						"commercetools_shipping_method.standard", "localized_description.en", newDescription,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "is_default", "true",
					),
					resource.TestCheckResourceAttrSet(
						"commercetools_shipping_method.standard", "tax_category_id",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_method.standard", "predicate", newPredicate,
					),
				),
			},
		},
	})
}

func testAccShippingMethodConfig(name string, key string, description string, localizedDescription string, isDefault bool, setTaxCategory bool, predicate string) string {
	taxCategoryReference := ""
	if setTaxCategory {
		taxCategoryReference = "tax_category_id = \"${commercetools_tax_category.test.id}\""
	}
	return fmt.Sprintf(`
resource "commercetools_tax_category" "test" {
	name = "test"
	key = "test"
	description = "test"
}

resource "commercetools_shipping_method" "standard" {
	name = "%s"
	key = "%s"
	description = "%s"
	localized_description = {
		en = "%s"
	}
	is_default = "%t"
	predicate = "%s"

	%s
	`, name, key, description, localizedDescription, isDefault, predicate, taxCategoryReference) + "\n}\n"
}

func testAccCheckShippingMethodDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_tax_category":
			{
				response, err := client.TaxCategories().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
					}
					continue
				}

				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		case "commercetools_shipping_method":
			{
				response, err := client.ShippingMethods().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("shipping method (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		default:
			continue
		}
	}
	return nil
}
