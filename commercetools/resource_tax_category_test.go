package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTaxCategoryCreate_basic(t *testing.T) {
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxCategoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "name", "Testing tax category",
					),
				),
			},
		},
	})
}

func testAccCheckTaxCategoryDestroy(s *terraform.State) error {
	return nil
}

func testAccTaxCategoryConfig(name string) string {
	return `resource "commercetools_tax_category" "standard" {
	name = "Testing tax category"
	rate {
		name = "19% MwSt"
		amount = 0.19
		included_in_price = false
		country = "DE"
	}
	rate {
		name = "21% BTW"
		amount = 0.21
		country = "NL"
		included_in_price = false
	}
	rate {
		name = "5% US"
		amount = 0.05
		country = "US"
		included_in_price = true
	}
	rate {
		name = "0% VAT"
		amount = 0.0
		included_in_price = true
		country = "GB"
	}
	}`
}
