package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDiscountCodeCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDiscountCodeConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "name.en", "Standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "code", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_from", "2020-01-02T15:04:05.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_until", "2021-01-02T15:04:05.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "is_active", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "max_applications_per_customer", "10",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "max_applications", "100",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.0", "cart-discount-id-0",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.1", "cart-discount-id-1",
					),
				),
			},
			{
				Config: testAccDiscountCodeUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "name.en", "Standard name new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "description.en", "Standard description new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "code", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_from", "2018-01-02T15:04:05.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_until", "2019-01-02T15:04:05.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "is_active", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "predicate", "1=2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "max_applications_per_customer", "5",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "max_applications", "50",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.0", "cart-discount-id-0",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.1",
					),
				),
			},
			{
				Config: testAccDiscountCodeRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"commercetools_discount_code.standard", "name.en",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_discount_code.standard", "description.en",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "code", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_from", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_until", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "is_active", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "predicate", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "max_applications_per_customer", "0",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "max_applications", "0",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.0", "cart-discount-id-0",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.1",
					),
				),
			},
		},
	})
}

func testAccDiscountCodeConfig() string {
	return `
	resource "commercetools_discount_code" "standard" {
		name = {
		  en = "Standard name"
		}
		description = {
			en = "Standard description"
		  }
		code           = "2"
		valid_from             = "2020-01-02T15:04:05.000Z"
		valid_until            = "2021-01-02T15:04:05.000Z"
		is_active      = true
        predicate      = "1=1"
        max_applications_per_customer = 10
        max_applications    = 100
		cart_discounts = ["cart-discount-id-0", "cart-discount-id-1"]
	  }  `
}

func testAccDiscountCodeUpdate() string {
	return `
	resource "commercetools_discount_code" "standard" {
		name = {
		  en = "Standard name new"
		}
		description = {
			en = "Standard description new"
		  }
		code           = "2"
		valid_from             = "2018-01-02T15:04:05.000Z"
		valid_until            = "2019-01-02T15:04:05.000Z"
		is_active      = false
        predicate      = "1=2"
        max_applications_per_customer = 5
        max_applications    = 50
		cart_discounts = ["cart-discount-id-0"]
	  }  `
}

func testAccDiscountCodeRemoveProperties() string {
	return `
	resource "commercetools_discount_code" "standard" {
		code           = "2"
		cart_discounts = ["cart-discount-id-0"]
	  }  `
}
