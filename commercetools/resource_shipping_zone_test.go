package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccShippingZone_createAndUpdate(t *testing.T) {

	name := "name"
	description := "description"

	newName := "new name"
	newDescription := "new description"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneConfig(name, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "description", description,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.#", "2",
					),
				),
			},
			{
				Config: testAccShippingZoneConfig(newName, newDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "name", newName,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "description", newDescription,
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.#", "2",
					),
				),
			},
		},
	})
}

func testAccShippingZoneConfig(name string, description string) string {
	return fmt.Sprintf(`
resource "commercetools_shipping_zone" "standard" {
	name = "%s"
	description = "%s"
	location = {
		country = "DE"
	}
	location = {
		country = "US"
		state = "Nevada"
	}
}`, name, description)
}

func testAccCheckShippingZoneDestroy(s *terraform.State) error {
	return nil
}
