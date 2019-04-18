package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccShippingMethod_createAndUpdate(t *testing.T) {

	name := "test method"
	key := "test-method"
	description := "test shipping method description"

	newName := "new test method"
	newKey := "new-test-method"
	newDescription := "new test shipping method description"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodConfig(name, key, description, false),
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
				),
			},
			{
				Config: testAccShippingMethodConfig(newName, newKey, newDescription, true),
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
				),
			},
		},
	})
}

func testAccShippingMethodConfig(name string, key string, description string, isDefault bool) string {
	return fmt.Sprintf(`
resource "commercetools_shipping_method" "standard" {
	name = "%s"
	key = "%s"
	description = "%s"
	is_default = "%t"
}`, name, key, description, isDefault)
}

func testAccCheckShippingMethodDestroy(s *terraform.State) error {
	return nil
}
