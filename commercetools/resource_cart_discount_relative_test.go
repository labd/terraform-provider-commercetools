package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/labd/commercetools-go-sdk/platform"

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

func TestAccCartDiscountRelative_CustomField(t *testing.T) {
	identifier := "relative_with_custom_field"
	resourceName := "commercetools_cart_discount." + identifier

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCartDiscountRelativeConfigWithCustomField(identifier, "relative_new_with_custom_field"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "relative_new_with_custom_field"),
					func(s *terraform.State) error {
						result, err := testGetCartDiscount(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Custom)
						assert.NotNil(t, result.Custom.Fields)
						assert.EqualValues(t, "foobar", result.Custom.Fields["my-string-field"])
						assert.EqualValues(t, []any{"ENUM-1", "ENUM-3"}, result.Custom.Fields["my-enum-set-field"])
						assert.EqualValues(t, map[string]interface{}{"centAmount": float64(150000), "currencyCode": "EUR", "fractionDigits": float64(2), "type": "centPrecision"}, result.Custom.Fields["my-money-field"])
						return nil
					},
				),
			},
			{
				Config: testAccCartDiscountRelativeUpdateCustomField(identifier, "relative_new_with_custom_field"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "relative_new_with_custom_field"),
					func(s *terraform.State) error {
						result, err := testGetCartDiscount(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Custom)
						assert.NotNil(t, result.Custom.Fields)
						assert.EqualValues(t, "foobar_foobar", result.Custom.Fields["my-string-field"])
						assert.EqualValues(t, []any{"ENUM-2"}, result.Custom.Fields["my-enum-set-field"])
						assert.EqualValues(t, map[string]interface{}{"centAmount": float64(2000), "currencyCode": "USD", "fractionDigits": float64(2), "type": "centPrecision"}, result.Custom.Fields["my-money-field"])
						return nil
					},
				),
			},
			{
				Config: testAccCartDiscountRelativeRemoveCustomField(identifier, "relative_new_with_custom_field"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "relative_new_with_custom_field"),
					func(s *terraform.State) error {
						result, err := testGetCartDiscount(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.Nil(t, result.Custom.Fields["my-string-field"])
						return nil
					},
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

func testAccCartDiscountRelativeConfigWithCustomField(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_type" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en = "for relative cart-discount"
			}
			description = {
				en = "Custom Field for relative cart-discount resource"
			}

			resource_type_ids = ["cart-discount"]

			field {
				name = "my-string-field"
				label = {
					en = "My Custom 'String' field"
				}
				type {
					name = "String"
				}
			}

			field {
				name = "my-enum-set-field"
				label = {
					en = "My Custom 'Set of enums' field"
				}
				type {
					name = "Set"
					element_type {
						name = "Enum"
						value {
							key   = "ENUM-1"
							label = "ENUM 1"
						}
						value {
							key   = "ENUM-2"
							label = "ENUM 2"
						}
						value {
							key   = "ENUM-3"
							label = "ENUM 3"
						}
					}
				}
			}

			field {
				name = "my-money-field"
				label = {
					en = "My Custom 'Money' field"
				}
				required = true
				type {
					name = "Money"
				}
				input_hint = "SingleLine"
			}
		}

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

			custom {
				type_id = commercetools_type.{{ .identifier }}.id
				fields = {
					"my-string-field" = "foobar"
					"my-enum-set-field" = jsonencode(["ENUM-1", "ENUM-3"])
					"my-money-field" = jsonencode({
						"type" : "centPrecision",
						"currencyCode" : "EUR",
						"centAmount" : 150000,
						"fractionDigits" : 2
					})
				}
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

func testAccCartDiscountRelativeUpdateCustomField(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_type" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en = "for relative cart-discount"
			}
			description = {
				en = "Custom Field for relative cart-discount resource"
			}

			resource_type_ids = ["cart-discount"]

			field {
				name = "my-string-field"
				label = {
					en = "My Custom 'String' field"
				}
				type {
					name = "String"
				}
			}

			field {
				name = "my-enum-set-field"
				label = {
					en = "My Custom 'Set of enums' field"
				}
				type {
					name = "Set"
					element_type {
						name = "Enum"
						value {
							key   = "ENUM-1"
							label = "ENUM 1"
						}
						value {
							key   = "ENUM-2"
							label = "ENUM 2"
						}
						value {
							key   = "ENUM-3"
							label = "ENUM 3"
						}
					}
				}
			}

			field {
				name = "my-money-field"
				label = {
					en = "My Custom 'Money' field"
				}
				required = true
				type {
					name = "Money"
				}
				input_hint = "SingleLine"
			}
		}
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

			custom {
				type_id = commercetools_type.{{ .identifier }}.id
				fields = {
					"my-string-field" = "foobar_foobar"
					"my-enum-set-field" = jsonencode(["ENUM-2"])
					"my-money-field" = jsonencode({
						"type" : "centPrecision",
						"currencyCode" : "USD",
						"centAmount" : 2000,
						"fractionDigits" : 2
					})
				}
			}
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

func testAccCartDiscountRelativeRemoveCustomField(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_type" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en = "for relative cart-discount"
			}
			description = {
				en = "Custom Field for relative cart-discount resource"
			}

			resource_type_ids = ["cart-discount"]

			field {
				name = "my-string-field"
				label = {
					en = "My Custom 'String' field"
				}
				type {
					name = "String"
				}
			}

			field {
				name = "my-enum-set-field"
				label = {
					en = "My Custom 'Set of enums' field"
				}
				type {
					name = "Set"
					element_type {
						name = "Enum"
						value {
							key   = "ENUM-1"
							label = "ENUM 1"
						}
						value {
							key   = "ENUM-2"
							label = "ENUM 2"
						}
						value {
							key   = "ENUM-3"
							label = "ENUM 3"
						}
					}
				}
			}

			field {
				name = "my-money-field"
				label = {
					en = "My Custom 'Money' field"
				}
				required = true
				type {
					name = "Money"
				}
				input_hint = "SingleLine"
			}
		}
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

			custom {
				type_id = commercetools_type.{{ .identifier }}.id
				fields = {
					"my-enum-set-field" = jsonencode(["ENUM-1", "ENUM-3"])
					"my-money-field" = jsonencode({
						"type" : "centPrecision",
						"currencyCode" : "USD",
						"centAmount" : 2000,
						"fractionDigits" : 2
					})
				}
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
