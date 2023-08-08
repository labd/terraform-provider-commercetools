package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCartDiscountFixed(t *testing.T) {
	identifier := "fixed"
	resourceName := "commercetools_cart_discount.fixed"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountFixedConfig(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "fixed name"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "fixed"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.cent_amount", "1000"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.cent_amount", "2000"),
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "shipping"),
				),
			},
			{
				Config: testAccCartDiscountFixedMultiBuyConfig(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "fixed name"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "fixed"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.cent_amount", "1000"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.cent_amount", "2000"),
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "multiBuyLineItems"),
					resource.TestCheckResourceAttr(resourceName, "target.0.predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.trigger_quantity", "2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.discounted_quantity", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.max_occurrence", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "Cheapest"),
				),
			},
			{
				Config: testAccCartDiscountFixedMultiBuyCustomConfig(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "fixed name"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "fixed"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.0.cent_amount", "1000"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "value.0.money.1.cent_amount", "2000"),
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "multiBuyCustomLineItems"),
					resource.TestCheckResourceAttr(resourceName, "target.0.predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.trigger_quantity", "2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.discounted_quantity", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.max_occurrence", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "MostExpensive"),
				),
			},
		},
	})
}

func testAccCartDiscountFixedConfig(identifier string) string {
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

func testAccCartDiscountFixedMultiBuyConfig(identifier string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			name = {
				en = "fixed name"
			}
			sort_order             = "0.9"
			predicate              = "1=1"

			target {
				type      			= "multiBuyLineItems"
				predicate      		= "1=1"
				trigger_quantity    = "2"
				discounted_quantity	= "1"
				max_occurrence      = "1"
				selection_mode 		= "Cheapest"
			}

			value {
				type      = "fixed"
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

func testAccCartDiscountFixedMultiBuyCustomConfig(identifier string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			name = {
				en = "fixed name"
			}
			sort_order             = "0.9"
			predicate              = "1=1"

			target {
				type      			= "multiBuyCustomLineItems"
				predicate      		= "1=1"
				trigger_quantity    = "2"
				discounted_quantity	= "1"
				max_occurrence      = "1"
				selection_mode 		= "MostExpensive"
			}

			value {
				type      = "fixed"
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
