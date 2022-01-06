package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCartDiscountCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "key", "standard",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "sort_order", "0.9",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "stacking_mode", "Stacking",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "requires_discount_code", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "valid_from", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "valid_until", "2019-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.0.type", "lineItems",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.0.predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "value.0.permyriad", "1000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "is_active", "true",
					),
				),
			},
			{
				Config: testAccCartDiscountUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "description.en", "Standard description new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "sort_order", "0.8",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "stacking_mode", "Stacking",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "requires_discount_code", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "valid_from", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "valid_until", "2019-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.0.type", "lineItems",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.0.predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "value.0.permyriad", "1000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "is_active", "false",
					),
				),
			},
			{
				Config: testAccCartDiscountRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_cart_discount.standard", "description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "sort_order", "0.8",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "stacking_mode", "Stacking",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "requires_discount_code", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "valid_from", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "valid_until", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.0.type", "lineItems",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "target.0.predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "value.0.permyriad", "1000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_cart_discount.standard", "is_active", "true",
					),
				),
			},
		},
	})
}

func testAccCartDiscountConfig() string {
	return `
	resource "commercetools_cart_discount" "standard" {
		key = "standard"
		name = {
		  en = "standard name"
		}
		description = {
			en = "Standard description"
		  }
		sort_order             = "0.9"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2018-01-02T15:04:05Z"
		valid_until            = "2019-01-02T15:04:05Z"
		target {
		  type      = "lineItems"
		  predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }
	  `
}

func testAccCartDiscountUpdate() string {
	return `
	resource "commercetools_cart_discount" "standard" {
		key = "standard_new"
		name = {
		  en = "standard name"
		}
		description = {
			en = "Standard description new"
		  }
		sort_order             = "0.8"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2018-01-02T15:04:05Z"
		valid_until            = "2019-01-02T15:04:05Z"
		target {
			type      = "lineItems"
			predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}

		is_active = false
	  }
	  `
}

func testAccCartDiscountRemoveProperties() string {
	return `
	resource "commercetools_cart_discount" "standard" {
		key = "standard_new"
		name = {
		  en = "standard name"
		}
		sort_order             = "0.8"
		predicate              = "1=1"
		requires_discount_code = true
		target {
			type      = "lineItems"
			predicate = "1=1"
		}
		value {
			type      = "relative"
			permyriad = 1000
		}
	  }
	  `
}

func testAccCheckCartDiscountDestroy(s *terraform.State) error {
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
