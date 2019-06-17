package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccShippingZoneRate_create(t *testing.T) {

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
	}
`, taxCategoryName, shippingMethodName, currencyCode)
}

func testAccCheckShippingZoneRateDestroy(s *terraform.State) error {
	return nil
}
