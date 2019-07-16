package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccProductDiscountCreate_basic(t *testing.T) {
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProductDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDiscountConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.tf_test_discount", "is_active", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.tf_test_discount", "sort_order", "0.1234",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.tf_test_discount", "value.#", "1",
					),
				),
			},
		},
	})
}

func testAccCheckProductDiscountDestroy(s *terraform.State) error {
	return nil
}

func testAccProductDiscountConfig(name string) string {
	return fmt.Sprintf(`
resource "commercetools_product_discount" "tf_test_discount" {
	name = {
		en = "%[1]s"
	}
	is_active = false
	sort_order = "0.1234"
	value {
		type      = "relative"
		permyriad = 1000
	}
}`, name)
}
