package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccShippingZoneRate_createAndUpdate(t *testing.T) {

	taxCategoryName := acctest.RandomWithPrefix("tf-acc-test")
	shippingMethodName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "commercetools_shipping_zone_rate.standard-de"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingZoneRateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneRateConfig(taxCategoryName, shippingMethodName, "EUR"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "price.0.cent_amount", "5000"),
					resource.TestCheckResourceAttr(resourceName, "price.0.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "free_above.0.cent_amount", "50000"),
					resource.TestCheckResourceAttr(resourceName, "free_above.0.currency_code", "EUR"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.0.type", "CartValue"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.0.minimum_cent_amount", "5000"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.0.price.0.cent_amount", "5000"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.1.type", "CartValue"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.1.minimum_cent_amount", "20000"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.1.price.0.cent_amount", "2000"),
				),
			},
			{
				Config: testAccShippingZoneRateUpdate(taxCategoryName, shippingMethodName, "USD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "price.0.cent_amount", "4321"),
					resource.TestCheckResourceAttr(resourceName, "price.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "free_above.0.cent_amount", "12345"),
					resource.TestCheckResourceAttr(resourceName, "free_above.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.0.type", "CartScore"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.0.score", "10"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.0.price.0.cent_amount", "5000"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.1.type", "CartScore"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.1.score", "20"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.1.price.0.cent_amount", "2000"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.2.type", "CartScore"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.2.score", "30"),
					resource.TestCheckResourceAttr(resourceName, "shipping_rate_price_tier.2.price_function.0.function", "x + 1"),
				),
			},
		},
	})
}

func testAccShippingZoneRateConfig(taxCategoryName string, shippingMethodName string, currencyCode string) string {
	return hclTemplate(`
		resource "commercetools_tax_category" "standard" {
			name        = "{{ .taxCategoryName }}"
			key         = "{{ .taxCategoryName }}"
			description = "Terraform test rate tax"
		}

		resource "commercetools_shipping_method" "standard" {
			name            = "{{ .shippingMethodName }}"
			key             = "{{ .shippingMethodName }}"
			description     = "Terraform test tax category"
			tax_category_id = commercetools_tax_category.standard.id
			predicate		= "1 = 1"
		}

		resource "commercetools_shipping_zone" "de" {
			name        = "DE"
			description = "Germany"
			location {
				country = "DE"
			}
		}

		resource "commercetools_shipping_zone_rate" "standard-de" {
			shipping_method_id = commercetools_shipping_method.standard.id
			shipping_zone_id   = commercetools_shipping_zone.de.id

			price {
				cent_amount   = 5000
				currency_code    = "{{ .currencyCode }}"
			}

			free_above {
				cent_amount   = 50000
				currency_code    = "{{ .currencyCode }}"
			}

			shipping_rate_price_tier {
				type                = "CartValue"
				minimum_cent_amount = 5000

				price {
					cent_amount      = 5000
					currency_code    = "{{ .currencyCode }}"
				}
			}

			shipping_rate_price_tier {
				type                = "CartValue"
				minimum_cent_amount = 20000

				price {
					cent_amount      = 2000
					currency_code    = "{{ .currencyCode }}"
				}
			}
		}`,
		map[string]any{
			"taxCategoryName":    taxCategoryName,
			"shippingMethodName": shippingMethodName,
			"currencyCode":       currencyCode,
		})
}

func testAccShippingZoneRateUpdate(taxCategoryName string, shippingMethodName string, currencyCode string) string {
	return hclTemplate(`
		resource "commercetools_tax_category" "standard" {
			name        = "{{ .taxCategoryName }}"
			key         = "{{ .taxCategoryName }}"
			description = "Terraform test rate tax"
		}

		resource "commercetools_shipping_method" "standard" {
			name            = "{{ .shippingMethodName }}"
			key             = "{{ .shippingMethodName }}"
			description     = "Terraform test tax category"
			tax_category_id = commercetools_tax_category.standard.id
			predicate		= "1 = 1"
		}

		resource "commercetools_shipping_zone" "de" {
			name        = "DE"
			description = "Germany"
			location {
				country = "DE"
			}
		}

		resource "commercetools_shipping_zone_rate" "standard-de" {
			shipping_method_id = commercetools_shipping_method.standard.id
			shipping_zone_id   = commercetools_shipping_zone.de.id

			price {
				cent_amount   = 4321
				currency_code    = "{{ .currencyCode }}"
			}

			free_above {
				cent_amount   = 12345
				currency_code    = "{{ .currencyCode }}"
			}

			shipping_rate_price_tier {
				type                = "CartScore"
				score               = 10

				price {
					cent_amount      = 5000
					currency_code    = "{{ .currencyCode }}"
				}
			}

			shipping_rate_price_tier {
				type                = "CartScore"
				score               = 20

				price {
					cent_amount      = 2000
					currency_code    = "{{ .currencyCode }}"
				}
			}
			shipping_rate_price_tier {
				type  = "CartScore"
				score = 30

				price_function {
					function      = "x + 1"
					currency_code = "{{ .currencyCode }}"
				}
			}
		}`,
		map[string]any{
			"taxCategoryName":    taxCategoryName,
			"shippingMethodName": shippingMethodName,
			"currencyCode":       currencyCode,
		})
}

func testAccCheckShippingZoneRateDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())
	// TODO: Do we want to check trailing rates separately? Similar to resource_tax_category_test_rate.go

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_tax_category":
			{
				response, err := client.TaxCategories().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		case "commercetools_shipping_method":
			{
				response, err := client.TaxCategories().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {

						return fmt.Errorf("shipping method (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		case "commercetools_shipping_zone":
			{
				response, err := client.Zones().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("shipping zone (%s) still exists", rs.Primary.ID)
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
