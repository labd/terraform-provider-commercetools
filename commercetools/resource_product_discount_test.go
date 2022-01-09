package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccProductDiscountCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProductDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductDiscountConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "2021-01-01T00:00:00.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "2022-01-01T00:00:00.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.type", "absolute",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.money.currency_code", "EUR",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.money.cent_amount", "50",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "is_active", "false",
					),
				),
			},
			{
				Config: testAccProductDiscountUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "description.en", "Standard description new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.8",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "2021-01-01T00:00:00.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "2022-01-01T00:00:00.000Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.permyriad", "1000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "is_active", "true",
					),
				),
			},
		},
	})
}

func testAccProductDiscountConfig() string {
	return `
	resource "commercetools_product_discount" "standard" {
	  name        = {
		en = "standard name new"
	  }
	  key         = "standard_new"
	  description = {
		en = "Standard description new"
	  }
	  predicate   = "1=1"
	  sort_order  = "0.1"
	  valid_from  = "2021-01-01T00:00:00.000Z"
	  valid_until = "2022-01-01T00:00:00.000Z"
	  value {
		type = "absolute"
		money {
		  currency_code = "EUR"
		  cent_amount   = 50
		}
	  }
	}
	`
}

func testAccProductDiscountUpdate() string {
	return `
	resource "commercetools_product_discount" "standard" {
	  name        = {
		en = "My new product discount name"
	  }
	  key         = "standard_new"
	  description = {
		en = "My new product discount description"
	  }
	  predicate   = "1=1"
	  sort_order  = "0.8"
	  is_active   = true
	  value {
		type = "relative"
		permyriad = 1000
	  }
	}
  `
}

func testAccCheckProductDiscountDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product_discount" {
			continue
		}
		response, err := client.ProductDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product discount (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
