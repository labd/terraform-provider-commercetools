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

func TestAccCartDiscountCreate_basic(t *testing.T) {
	identifier := "standard"
	resourceName := "commercetools_cart_discount.standard"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountConfig(identifier, "standard"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "standard"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "standard name"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "Standard description"),
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.9"),
					resource.TestCheckResourceAttr(resourceName, "predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "stacking_mode", "Stacking"),
					resource.TestCheckResourceAttr(resourceName, "requires_discount_code", "true"),
					resource.TestCheckResourceAttr(resourceName, "valid_from", "2018-01-02T15:04:05Z"),
					resource.TestCheckResourceAttr(resourceName, "valid_until", "2019-01-02T15:04:05Z"),
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "lineItems"),
					resource.TestCheckResourceAttr(resourceName, "target.0.predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "relative"),
					resource.TestCheckResourceAttr(resourceName, "value.0.permyriad", "1000"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "true"),
					func(s *terraform.State) error {
						res, err := testGetCartDiscount(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.EqualValues(t, res.Name["en"], "standard name")
						assert.EqualValues(t, (*res.Key), "standard")
						assert.True(t, res.IsActive)
						assert.True(t, res.RequiresDiscountCode)
						return nil
					},
				),
			},
			{
				Config: testAccCartDiscountUpdate(identifier, "standard_new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "standard_new"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "standard name"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "Standard description new"),
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.8"),
					resource.TestCheckResourceAttr(resourceName, "predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "stacking_mode", "Stacking"),
					resource.TestCheckResourceAttr(resourceName, "requires_discount_code", "true"),
					resource.TestCheckResourceAttr(resourceName, "valid_from", "2018-01-02T15:04:05Z"),
					resource.TestCheckResourceAttr(resourceName, "valid_until", "2019-01-02T15:04:05Z"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "lineItems"),
					resource.TestCheckResourceAttr(resourceName, "target.0.predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "relative"),
					resource.TestCheckResourceAttr(resourceName, "value.0.permyriad", "1000"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "false"),
				),
			},
			{
				Config: testAccCartDiscountRemoveProperties(identifier, "standard_new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "standard_new"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "standard name"),
					resource.TestCheckNoResourceAttr(resourceName, "description.en"),
					resource.TestCheckResourceAttr(resourceName, "sort_order", "0.8"),
					resource.TestCheckResourceAttr(resourceName, "predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "stacking_mode", "Stacking"),
					resource.TestCheckResourceAttr(resourceName, "requires_discount_code", "true"),
					resource.TestCheckResourceAttr(resourceName, "valid_from", ""),
					resource.TestCheckResourceAttr(resourceName, "valid_until", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.type", "lineItems"),
					resource.TestCheckResourceAttr(resourceName, "target.0.predicate", "1=1"),
					resource.TestCheckResourceAttr(resourceName, "value.0.type", "relative"),
					resource.TestCheckResourceAttr(resourceName, "value.0.permyriad", "1000"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "true"),
				),
			},
		},
	})
}

func testAccCartDiscountConfig(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			key = "{{ .key }}"
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
	`, map[string]any{
		"identifier": identifier,
		"key":        key,
	})
}

func testAccCartDiscountUpdate(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			key = "{{ .key }}"
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
	`, map[string]any{
		"identifier": identifier,
		"key":        key,
	})
}

func testAccCartDiscountRemoveProperties(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			key = "{{ .key }}"
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
	`, map[string]any{
		"identifier": identifier,
		"key":        key,
	})
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

func testGetCartDiscount(s *terraform.State, identifier string) (*platform.CartDiscount, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Cart Discount %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.CartDiscounts().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
