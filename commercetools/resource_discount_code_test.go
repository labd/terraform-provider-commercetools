package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestAccDiscountCodeCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDiscountCodeDestroy,
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
						"commercetools_discount_code.standard", "valid_from", "2020-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_until", "2021-01-02T15:04:05Z",
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
						"commercetools_discount_code.standard", "valid_from", "2018-01-02T15:04:05Z",
					),
					resource.TestCheckResourceAttr(
						"commercetools_discount_code.standard", "valid_until", "2019-01-02T15:04:05Z",
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
					resource.TestCheckNoResourceAttr(
						"commercetools_discount_code.standard", "cart_discounts.1",
					),
				),
			},
		},
	})
}

func TestAccDiscountCode_CustomField(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCustomerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiscountCodeCustomField(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						res, err := testGetDiscountCode(s, "commercetools_discount_code.standard")
						if err != nil {
							return err
						}
						assert.NotNil(t, res)
						assert.NotNil(t, res.Custom)
						return nil
					},
				),
			},
		},
	})
}

func testAccDiscountCodeConfig() string {
	return `
	resource "commercetools_cart_discount" "standard" {
		key = "test_key"
		name = {
			en = "best cart discount"
		}
		description = {
			en = "Standard description"
		}
		sort_order             = "0.9123"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2020-01-02T15:04:05Z"
		valid_until            = "2021-01-02T15:04:05Z"

		target {
			type      = "lineItems"
			predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }

	resource "commercetools_cart_discount" "standard_2" {
		key = "another_test_key"
		name = {
			en = "best cart discount the second"
		}
		description = {
			en = "Standard description"
		}
		sort_order             = "0.9321"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2020-01-02T15:04:05Z"
		valid_until            = "2021-01-02T15:04:05Z"

		target {
			type      = "lineItems"
			predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }

	resource "commercetools_discount_code" "standard" {
		name = {
			en = "Standard name"
		}
		description = {
			en = "Standard description"
		}
		code        = "2"
		valid_from  = "2020-01-02T15:04:05Z"
		valid_until = "2021-01-02T15:04:05Z"
		is_active   = true
        predicate   = "1=1"

        max_applications_per_customer = 10
        max_applications              = 100

		cart_discounts = [commercetools_cart_discount.standard.id, commercetools_cart_discount.standard_2.id]
	  }`
}

func testAccDiscountCodeUpdate() string {
	return `
	resource "commercetools_cart_discount" "standard" {
		key = "test_key"
		name = {
			en = "best cart discount"
		}
		description = {
			en = "Standard description"
		}
		sort_order             = "0.9123"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2020-01-02T15:04:05Z"
		valid_until            = "2021-01-02T15:04:05Z"
		target {
		  type      = "lineItems"
		  predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	}

	resource "commercetools_cart_discount" "standard_2" {
		key = "another_test_key"
		name = {
			en = "best cart discount the second"
		}
		description = {
			en = "Standard description"
		}
		sort_order             = "0.9321"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2020-01-02T15:04:05Z"
		valid_until            = "2021-01-02T15:04:05Z"

		target {
			type      = "lineItems"
			predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	}

	resource "commercetools_discount_code" "standard" {
		name = {
			en = "Standard name new"
		}
		description = {
			en = "Standard description new"
		}
		code           = "2"
		valid_from     = "2018-01-02T15:04:05Z"
		valid_until    = "2019-01-02T15:04:05Z"
		is_active      = false
        predicate      = "1=2"

        max_applications_per_customer = 5
        max_applications              = 50

		cart_discounts = [commercetools_cart_discount.standard.id]
	  }  `
}

func testAccDiscountCodeRemoveProperties() string {
	return `
		resource "commercetools_cart_discount" "standard" {
		key = "test_key"
		name = {
			en = "best cart discount"
		}
		description = {
			en = "Standard description"
		}
		sort_order             = "0.9123"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2020-01-02T15:04:05Z"
		valid_until            = "2021-01-02T15:04:05Z"

		target {
		  type      = "lineItems"
		  predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }

	resource "commercetools_cart_discount" "standard_2" {
		key = "another_test_key"
		name = {
			en = "best cart discount the second"
		}
		description = {
			en = "Standard description"
		}
		sort_order             = "0.9321"
		predicate              = "1=1"
		stacking_mode          = "Stacking"
		requires_discount_code = true
		valid_from             = "2020-01-02T15:04:05Z"
		valid_until            = "2021-01-02T15:04:05Z"

		target {
		  type      = "lineItems"
		  predicate = "1=1"
		}

		value {
			type      = "relative"
			permyriad = 1000
		}
	  }

	resource "commercetools_discount_code" "standard" {
		code           = "2"
		cart_discounts = [commercetools_cart_discount.standard.id]
	  }`
}

func testAccDiscountCodeCustomField() string {
	return `
	resource "commercetools_type" "test" {
		key = "test-for-discount-code"
		name = {
			en = "for discount-code"
		}
		description = {
			en = "Custom Field for discount-code resource"
		}

		resource_type_ids = ["discount-code"]

		field {
			name = "my-field"
			label = {
				en = "My Custom field"
			}
			type {
				name = "String"
			}
		}
	}
	resource "commercetools_discount_code" "standard" {
		name = {
			en = "Standard name"
		}
		description = {
			en = "Standard description"
		}
		code        = "2"
		valid_from  = "2020-01-02T15:04:05Z"
		valid_until = "2021-01-02T15:04:05Z"
		is_active   = true
        predicate   = "1=1"

        max_applications_per_customer = 10
        max_applications              = 100

		cart_discounts = []

		custom {
			type_id = commercetools_type.test.id
			fields = {
				"my-field" = "foobar"
			}
		}
	}`

}

func testAccCheckDiscountCodeDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_cart_discount":
			{
				response, err := client.CartDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("cart discount (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		case "commercetools_discount_code":
			{
				response, err := client.CartDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("discount code (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		default:
			continue
		}
	}
	return nil
}

func testGetDiscountCode(s *terraform.State, identifier string) (*platform.DiscountCode, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("DiscountCode not found")
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.DiscountCodes().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
