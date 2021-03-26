package commercetools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/labd/commercetools-go-sdk/commercetools"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccShippingZone_createAndUpdateWithID(t *testing.T) {

	key := "key"
	name := "name"
	description := "description"

	newKey := "new-key"
	newName := "new name"
	newDescription := "new description"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneConfig(name, description, key),
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
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "key", key,
					),
				),
			},
			{
				Config: testAccShippingZoneConfig(newName, newDescription, newKey),
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
					resource.TestCheckResourceAttr(
						"commercetools_shipping_zone.standard", "key", newKey,
					),
				),
			},
		},
	})
}

func testAccShippingZoneConfig(name string, description string, key string) string {
	return fmt.Sprintf(`
resource "commercetools_shipping_zone" "standard" {
	name = "%s"
	description = "%s"
	key = "%s"
	location {
		country = "DE"
	}
	location {
		country = "US"
		state = "Nevada"
	}
}`, name, description, key)
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
				Config: testAccShippingZoneConfig(name, description, name),
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
	location {
		country = "DE"
	}
	location {
		country = "ES"
	}
	location {
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
				Config: testAccShippingZoneConfig(name, description, name),
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
	location {
		country = "US"
		state = "Nevada"
	}
}`
}

func testAccCheckShippingZoneDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*commercetools.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_shipping_zone" {
			continue
		}
		response, err := conn.ZoneGetWithID(context.Background(), rs.Primary.ID)
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("shipping zone (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
		if !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Not Found (404)") {
			return err
		}
	}
	return nil
}
