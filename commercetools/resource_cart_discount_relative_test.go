package commercetools

import (
	"context"
	"fmt"
	"github.com/labd/commercetools-go-sdk/platform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccCartDiscountRelative(t *testing.T) {
	identifier := "relative"
	resourceName := "commercetools_cart_discount.relative"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountRelativeConfig(identifier, "relative"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "relative"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "relative name"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "relative description"),
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
						assert.EqualValues(t, res.Name["en"], "relative name")
						assert.EqualValues(t, *res.Key, "relative")
						assert.True(t, res.IsActive)
						assert.True(t, res.RequiresDiscountCode)
						return nil
					},
				),
			},
			{
				Config: testAccCartDiscountRelativeUpdate(identifier, "relative_new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "relative_new"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "relative name"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "relative description new"),
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
				Config: testAccCartDiscountRelativeRemoveProperties(identifier, "relative_new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "relative_new"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "relative name"),
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

func testAccCartDiscountRelativeConfig(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en = "relative name"
			}
			description = {
				en = "relative description"
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

func testAccCartDiscountRelativeUpdate(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en = "relative name"
			}
			description = {
				en = "relative description new"
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

func testAccCartDiscountRelativeRemoveProperties(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_cart_discount" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en = "relative name"
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
