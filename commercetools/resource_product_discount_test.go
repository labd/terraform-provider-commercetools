package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
						"commercetools_product_discount.standard", "key", "standard",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.95",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "2019-01-02T15:04:05Z",
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
			{
				Config: testAccProductDiscountUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name",
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
						"commercetools_product_discount.standard", "valid_from", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "2019-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.type", "relative",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "value.0.permyriad", "1000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "is_active", "false",
					),
				),
			},
			{
				Config: testAccProductDiscountRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "key", "standard_new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "name.en", "standard name",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_product_discount.standard", "description.en",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "sort_order", "0.8",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "predicate", "1=1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_from", "",
					),
					resource.TestCheckResourceAttr(
						"commercetools_product_discount.standard", "valid_until", "",
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
	return fmt.Sprintf(`
	resource "commercetools_product_discount" "standard" {
		key = "standard"
		name = {
		  en = "standard name"
		}
		description = {
			en = "Standard description"
		  }
		predicate              = "1=1"
		sort_order             = "0.95"
		is_active              = true
		valid_from             = "2018-01-02T15:04:05Z"
		valid_until            = "2019-01-02T15:04:05Z"

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }
	  `)
}

func testAccProductDiscountUpdate() string {
	return `
	resource "commercetools_product_discount" "standard" {
		key = "standard_new"
		name = {
		  en = "standard name"
		}
		description = {
			en = "Standard description new"
		  }
		sort_order             = "0.8"
		predicate              = "1=1"
		valid_from             = "2018-01-02T15:04:05Z"
		valid_until            = "2019-01-02T15:04:05Z"

		value {
			type      = "relative"
			permyriad = 1000
		}

		is_active = false
	  }
	  `
}

func testAccProductDiscountRemoveProperties() string {
	return `
	resource "commercetools_product_discount" "standard" {
		key = "standard_new"
		name = {
		  en = "standard name"
		}
		sort_order             = "0.8"
		predicate              = "1=1"
		value {
			type      = "relative"
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
