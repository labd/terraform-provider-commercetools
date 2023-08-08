package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCartDiscountGiftLineItem(t *testing.T) {
	identifier := "gift_line_item"
	resourceName := "commercetools_cart_discount.gift_line_item"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountGiftLineItemConfig(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "giftLineItem name"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "giftLineItem"),
					resource.TestCheckResourceAttr(resourceName, "value.0.product_id", "product-id"),
					resource.TestCheckResourceAttr(resourceName, "value.0.variant_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "value.0.supply_channel_id", "supply-channel-id"),
					resource.TestCheckResourceAttr(resourceName, "value.0.distribution_channel_id", "distribution-channel-id"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "0"),
				),
			},
		},
	})
}

func testAccCartDiscountGiftLineItemConfig(identifier string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			name = {
				en = "giftLineItem name"
			}
			sort_order             = "0.9"
			predicate              = "1=1"

			value {
				type                    = "giftLineItem"
				product_id              = "product-id"
				variant_id              = 1
				supply_channel_id       = "supply-channel-id"
				distribution_channel_id = "distribution-channel-id"
			}
		}
	`, map[string]any{
		"identifier": identifier,
	})
}
