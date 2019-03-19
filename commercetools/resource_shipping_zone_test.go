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

func TestAccShippingZone_createAndAddLocation(t *testing.T) {

	name := "name"
	description := "description"

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
						"commercetools_shipping_zone.standard", "location.0.country", "DE",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.1.country", "US",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.1.state", "Nevada",
					),
				),
			},
			{
				Config: testAccShippingZoneConfigLocationAdded(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "name", "the zone",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "description", "the description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.0.country", "DE",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.1.country", "ES",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.2.country", "US",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.2.state", "Nevada",
					),
				),
			},
		},
	})
}

func testAccShippingZoneConfigLocationAdded() string {
	return `
resource "commercetools_shipping_zone" "standard" {
	name = "the zone"
	description = "the description"
	location = {
		country = "DE"
	}
	location = {
		country = "ES"
	}
	location = {
		country = "US"
		state = "Nevada"
	}
}`
}

func TestAccShippingZone_createAndRemoveLocation(t *testing.T) {

	name := "name"
	description := "description"

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
						"commercetools_shipping_zone.standard", "location.0.country", "DE",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.1.country", "US",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.1.state", "Nevada",
					),
				),
			},
			{
				Config: testAccShippingZoneConfigLocationRemoved(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "name", "the zone",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "description", "the description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.0.country", "US",
					),
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "location.0.state", "Nevada",
					),
				),
			},
		},
	})
}

func testAccShippingZoneConfigLocationRemoved() string {
	return `
resource "commercetools_shipping_zone" "standard" {
	name = "the zone"
	description = "the description"
	location = {
		country = "US"
		state = "Nevada"
	}
}`
}

func testAccCheckShippingZoneDestroy(s *terraform.State) error {
	return nil
}
