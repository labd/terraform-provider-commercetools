package commercetools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/labd/commercetools-go-sdk/commercetools"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccShippingZoneRate_createAndUpdate(t *testing.T) {

	taxCategoryName := acctest.RandomWithPrefix("tf-acc-test")
	shippingMethodName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingZoneRateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneRateConfig(taxCategoryName, shippingMethodName, "EUR"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "price.0.cent_amount", "5000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "price.0.currency_code", "EUR",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "free_above.0.cent_amount", "50000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "free_above.0.currency_code", "EUR",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.#", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.0.type", "CartValue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.0.minimum_cent_amount", "5000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.0.price.0.cent_amount", "5000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.1.type", "CartValue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.1.minimum_cent_amount", "20000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.1.price.0.cent_amount", "2000",
					),
				),
			},
			{
				Config: testAccShippingZoneRateUpdate(taxCategoryName, shippingMethodName, "USD"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "price.0.cent_amount", "4321",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "price.0.currency_code", "USD",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "free_above.0.cent_amount", "12345",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "free_above.0.currency_code", "USD",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.#", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.0.type", "CartScore",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.0.score", "10",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.0.price.0.cent_amount", "5000",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.1.type", "CartScore",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.1.score", "20",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone_rate.standard-de", "shipping_rate_price_tier.1.price.0.cent_amount", "2000",
					),
				),
			},
		},
	})
}

func testAccShippingZoneRateConfig(taxCategoryName string, shippingMethodName string, currencyCode string) string {
	return fmt.Sprintf(`
	resource "commercetools_tax_category" "standard" {
		name        = "%[1]s"
		key         = "%[1]s"
		description = "Terraform test rate tax"
	}

	resource "commercetools_shipping_method" "standard" {
		name            = "%[2]s"
		key             = "%[2]s"
		description     = "Terraform test tax category"
		tax_category_id = "${commercetools_tax_category.standard.id}"
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
		shipping_method_id = "${commercetools_shipping_method.standard.id}"
		shipping_zone_id   = "${commercetools_shipping_zone.de.id}"

		price {
		    cent_amount   = 5000
		    currency_code = "%[3]s"
		}

		free_above {
		    cent_amount   = 50000
		    currency_code = "%[3]s"
		}
		
		shipping_rate_price_tier {
            type                = "CartValue"
            minimum_cent_amount = 5000

            price {
              cent_amount      = 5000
              currency_code    = "%[3]s"
            }
		}

		shipping_rate_price_tier {
			type                = "CartValue"
            minimum_cent_amount = 20000

            price {
              cent_amount      = 2000
              currency_code    = "%[3]s"
            }
		}

	}
`, taxCategoryName, shippingMethodName, currencyCode)
}

func testAccShippingZoneRateUpdate(taxCategoryName string, shippingMethodName string, currencyCode string) string {
	return fmt.Sprintf(`
	resource "commercetools_tax_category" "standard" {
		name        = "%[1]s"
		key         = "%[1]s"
		description = "Terraform test rate tax"
	}

	resource "commercetools_shipping_method" "standard" {
		name            = "%[2]s"
		key             = "%[2]s"
		description     = "Terraform test tax category"
		tax_category_id = "${commercetools_tax_category.standard.id}"
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
		shipping_method_id = "${commercetools_shipping_method.standard.id}"
		shipping_zone_id   = "${commercetools_shipping_zone.de.id}"

		price {
		    cent_amount   = 4321
		    currency_code = "%[3]s"
		}

		free_above {
		    cent_amount   = 12345
		    currency_code = "%[3]s"
		}
		
        shipping_rate_price_tier {
            type                = "CartScore"
            score               = 10

            price {
              cent_amount      = 5000
              currency_code    = "%[3]s"
            }
		}

		shipping_rate_price_tier {
			type                = "CartScore"
            score               = 20

            price {
              cent_amount      = 2000
              currency_code    = "%[3]s"
            }
		}

	}
`, taxCategoryName, shippingMethodName, currencyCode)
}

func testAccCheckShippingZoneRateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*commercetools.Client)
	// TODO: Do we want to check trailing rates separately? Similar to resource_tax_category_test_rate.go

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_tax_category":
			{
				response, err := conn.TaxCategoryGetWithID(context.Background(), rs.Primary.ID)
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
				if !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Not Found (404)") {
					return err
				}
			}
		case "commercetools_shipping_method":
			{
				response, err := conn.ShippingMethodGetWithID(context.Background(), rs.Primary.ID)
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {

						return fmt.Errorf("shipping method (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
				if !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Not Found (404)") {
					return err
				}
			}
		case "commercetools_shipping_zone":
			{
				response, err := conn.ZoneGetWithID(context.Background(), rs.Primary.ID)
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("shipping zone (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
				if !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Not Found (404)") {
					return err
				}
			}
		default:
			continue
		}
	}
	return nil
}
