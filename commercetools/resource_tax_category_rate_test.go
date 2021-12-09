package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccTaxCategoryRate_createAndUpdateWithID(t *testing.T) {

	name := acctest.RandomWithPrefix("tf-acc-test")
	amount := 0.2
	country := "DE"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTaxCategoryRateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxCategoryRateConfig(name, amount, true, country),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0.2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
				),
			},
			{
				Config: testAccTaxCategoryRateConfig(name, amount, false, country),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0.2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
				),
			},
			{
				Config: testAccTaxCategoryRateConfig(name, 0.0, true, country),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
				),
			},
		},
	})
}

func testAccTaxCategoryRateConfig(name string, amount float64, includedInPrice bool, country string) string {
	return fmt.Sprintf(`
resource "commercetools_tax_category" "standard" {
	name = "test-rate-category"
	key = "test-rate-category"
	description = "Test rate tax"
}

resource "commercetools_tax_category_rate" "test_rate" {
	tax_category_id = "${commercetools_tax_category.standard.id}"
	name = "%s"
	amount = %f
	included_in_price = %t
	country = "%s"
}
`, name, amount, includedInPrice, country)
}

func TestAccTaxCategoryRate_createAndUpdateSubRates(t *testing.T) {

	name := acctest.RandomWithPrefix("tf-acc-test")
	subRateAmount := 0.3
	amount := 0.2
	country := "DE"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTaxCategoryRateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxCategoryRateSubRatesConfig(name, subRateAmount, true, country, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0.3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "sub_rate.0.amount", "0.2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "sub_rate.0.name", "foo",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "sub_rate.1.amount", "0.1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "sub_rate.1.name", "foo2",
					),
				),
			},
			{
				Config: testAccTaxCategoryRateSubRatesConfig(name, amount, false, country, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0.2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
					resource.TestCheckNoResourceAttr("commercetools_tax_category_rate.test_rate", "sub_rate"),
				),
			},
		},
	})
}

func testAccTaxCategoryRateSubRatesConfig(name string, amount float64, includedInPrice bool, country string, addSubrates bool) string {
	if addSubrates {
		return fmt.Sprintf(`
resource "commercetools_tax_category" "standard" {
	name        = "test-rate-category"
	key         = "test-rate-category"
	description = "Test rate tax"
}

resource "commercetools_tax_category_rate" "test_rate" {
	tax_category_id = "${commercetools_tax_category.standard.id}"
	name              = "%s"
	amount            = %f
	included_in_price = %t
	country           = "%s"
	sub_rate {
		name = "foo"
		amount = 0.2
	}
	sub_rate {
		name = "foo2"
		amount = 0.1
	}
}
`, name, amount, includedInPrice, country)
	}
	return fmt.Sprintf(`
resource "commercetools_tax_category" "standard" {
	name        = "test-rate-category"
	key         = "test-rate-category"
	description = "Test rate tax"
}

resource "commercetools_tax_category_rate" "test_rate" {
	tax_category_id = "${commercetools_tax_category.standard.id}"
	name              = "%s"
	amount            = %f
	included_in_price = %t
	country           = "%s"
}
`, name, amount, includedInPrice, country)
}

func TestAccTaxCategoryRate_createAndUpdateBothRateAndTaxCategory(t *testing.T) {

	name := acctest.RandomWithPrefix("tf-acc-test")
	amount := 0.2
	country := "DE"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTaxCategoryRateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxCategoryRateDualUpdateConfig("foo", name, amount, true, country),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "description", "foo",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0.2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
				),
			},
			{
				Config: testAccTaxCategoryRateDualUpdateConfig("bar", name, amount, false, country),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "description", "bar",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "amount", "0.2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "included_in_price", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category_rate.test_rate", "country", country,
					),
				),
			},
		},
	})
}

func testAccTaxCategoryRateDualUpdateConfig(description string, name string, amount float64, includedInPrice bool, country string) string {
	return fmt.Sprintf(`
resource "commercetools_tax_category" "standard" {
	name = "test-rate-category"
	key = "test-rate-category"
	description = "%s"
}

resource "commercetools_tax_category_rate" "test_rate" {
	tax_category_id = "${commercetools_tax_category.standard.id}"
	name = "%s"
	amount = %f
	included_in_price = %t
	country = "%s"
}
`, description, name, amount, includedInPrice, country)
}

func testAccCheckTaxCategoryRateDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())
	var rateIDs []string
	// Because we can't directly query for Tax Categories, we are going to loop over the resources twice. Once to store
	// the tax rate IDs of any rates present, and once to check the categories and their rates
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_tax_category_rate" {
			continue
		}
		rateIDs = append(rateIDs, rs.Primary.ID)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_tax_category" {
			continue
		}
		response, err := client.TaxCategories().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && len(response.Rates) > 0 && response.ID == rs.Primary.ID {
				var trailingTestRates []string
				for _, rate := range response.Rates {
					if stringInSlice(*rate.ID, rateIDs) {
						trailingTestRates = append(trailingTestRates, *rate.ID)
					}
				}
				return fmt.Errorf("tax category %s still exists with rates (%v)", rs.Primary.ID, trailingTestRates)
			}
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
			}
			continue
		}

		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
