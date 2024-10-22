package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCartDiscountStores(t *testing.T) {
	identifier := "stores"
	resourceName := "commercetools_cart_discount.stores"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountWithoutStores(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "0"),
				),
			},
			{
				Config: testAccCartDiscountWithStores(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(resourceName, "stores.*", "my-store"),
				),
			},
			{
				Config: testAccCartDiscountWithoutStores(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "0"),
				),
			},
		},
	})
}

func testAccCartDiscountWithoutStores(identifier string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			name = {
				en = "fixed name"
			}
			sort_order             = "0.9"
			predicate              = "1=1"

			target {
				type      = "shipping"
			}

			value {
				type      = "fixed"
				money {
					currency_code = "USD"
					cent_amount   = 1000
				}
			}
		}
	`, map[string]any{
		"identifier": identifier,
	})
}

func testAccCartDiscountWithStores(identifier string) string {
	return hclTemplate(`
		resource "commercetools_store" "my-store-{{ .identifier }}" {
		  key = "my-store"
		  name = {
			en-US = "My store"
		  }
		  countries = ["NL", "BE"]
		  languages = ["nl-NL"]
		}
	
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			name = {
				en = "fixed name"
			}
  			stores = [commercetools_store.my-store-{{ .identifier }}.key]
			sort_order             = "0.9"
			predicate              = "1=1"

			target {
				type      = "shipping"
			}

			value {
				type      = "fixed"
				money {
					currency_code = "USD"
					cent_amount   = 1000
				}
			}
		}
	`, map[string]any{
		"identifier": identifier,
	})
}
