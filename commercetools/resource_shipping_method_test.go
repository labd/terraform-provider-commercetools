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
				Config: testAccShippingMethodConfig(name, key, description),
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
				),
			},
			{
				Config: testAccShippingMethodConfig(newName, newKey, newDescription),
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
				),
			},
		},
	})
}

func testAccShippingMethodConfig(name string, key string, description string) string {
	return fmt.Sprintf(`
resource "commercetools_shipping_method" "standard" {
	name = "%s"
	key = "%s"
	description = "%s"
}`, name, key, description)
}

func testAccCheckShippingMethodDestroy(s *terraform.State) error {
	return nil
}
