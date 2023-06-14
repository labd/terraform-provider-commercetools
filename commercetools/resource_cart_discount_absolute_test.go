package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCartDiscountCreate_absolute(t *testing.T) {
	identifier := "absolute"
	resourceName := "commercetools_cart_discount.absolute"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountAbsoluteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountAbsoluteConfig(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "absolute name"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "absolute"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.cent_amount", "1000"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.cent_amount", "2000"),
				),
			},
		},
	})
}

func testAccCartDiscountAbsoluteConfig(identifier string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			name = {
				en = "absolute name"
			}
			sort_order             = "0.9"
			predicate              = "1=1"

			target {
				type      = "lineItems"
				predicate = "1=1"
			}

			value {
				type      = "absolute"
				money {
					currency_code = "USD"
					cent_amount   = 1000
				}
				money {
					currency_code = "EUR"
					cent_amount   = 2000
				}
			}
		}
	`, map[string]any{
		"identifier": identifier,
	})
}

func testAccCheckCartDiscountAbsoluteDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_cart_discount" {
			continue
		}
		response, err := client.CartDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("cart discount (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
