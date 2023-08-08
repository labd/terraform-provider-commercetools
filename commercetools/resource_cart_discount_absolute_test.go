package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCartDiscountAbsolute(t *testing.T) {
	identifier := "absolute"
	resourceName := "commercetools_cart_discount.absolute"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
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
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "customLineItems"),
					resource.TestCheckResourceAttr(resourceName, "target.0.predicate", "1=1"),
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
				type      = "customLineItems"
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
