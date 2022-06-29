package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestExpandShippingZoneLocations(t *testing.T) {
	resource := resourceShippingZone().Schema["location"].Elem.(*schema.Resource)
	input := schema.NewSet(schema.HashResource(resource), []interface{}{
		map[string]interface{}{
			"country": "DE",
			"state":   "",
		},
		map[string]interface{}{
			"country": "US",
			"state":   "Nevada",
		},
	})
	actual := expandShippingZoneLocations(input)
	expected := []platform.Location{
		{
			Country: "DE",
			State:   nil,
		},
		{
			Country: "US",
			State:   stringRef("Nevada"),
		},
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestAccShippingZone_createAndUpdateWithID(t *testing.T) {

	key := "key"
	name := "name"
	description := "description"
	resourceName := "commercetools_shipping_zone.standard"

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
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "location.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "location.0.country", "DE"),
					resource.TestCheckResourceAttr(resourceName, "location.0.state", ""),
					resource.TestCheckResourceAttr(resourceName, "key", key),
				),
			},
			{
				Config: testAccShippingZoneConfig(newName, newDescription, newKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", newName),
					resource.TestCheckResourceAttr(resourceName, "description", newDescription),
					resource.TestCheckResourceAttr(resourceName, "location.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "key", newKey),
				),
			},
		},
	})
}

func testAccShippingZoneConfig(name, description, key string) string {
	return hclTemplate(`
		resource "commercetools_shipping_zone" "standard" {
			name        = "{{ .name }}"
			description = "{{ .description }}"
			key         = "{{ .key }}"

			location {
				country = "DE"
			}

			location {
				country = "US"
				state = "Nevada"
			}
		}`,
		map[string]any{
			"name":        name,
			"description": description,
			"key":         key,
		})
}

func TestAccShippingZone_createAndAddLocation(t *testing.T) {

	name := "name"
	description := "description"
	resourceName := "commercetools_shipping_zone.standard"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneConfig(name, description, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "location.0.country", "DE"),
					resource.TestCheckResourceAttr(resourceName, "location.1.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "location.1.state", "Nevada"),
				),
			},
			{
				Config: testAccShippingZoneConfigLocationAdded(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "the zone"),
					resource.TestCheckResourceAttr(resourceName, "description", "the description"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "location.0.country", "DE"),
					resource.TestCheckResourceAttr(resourceName, "location.1.country", "ES"),
					resource.TestCheckResourceAttr(resourceName, "location.2.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "location.2.state", "Nevada"),
				),
			},
		},
	})
}

func testAccShippingZoneConfigLocationAdded() string {
	return hclTemplate(`
		resource "commercetools_shipping_zone" "standard" {
			name        = "the zone"
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
		}`, map[string]any{})
}

func TestAccShippingZone_createAndRemoveLocation(t *testing.T) {

	name := "name"
	description := "description"
	resourceName := "commercetools_shipping_zone.standard"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneConfig(name, description, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "location.0.country", "DE"),
					resource.TestCheckResourceAttr(resourceName, "location.1.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "location.1.state", "Nevada"),
				),
			},
			{
				Config: testAccShippingZoneConfigLocationRemoved(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "the zone"),
					resource.TestCheckResourceAttr(resourceName, "description", "the description"),
					resource.TestCheckResourceAttr(resourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "location.0.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "location.0.state", "Nevada"),
				),
			},
		},
	})
}

func testAccShippingZoneConfigLocationRemoved() string {
	return hclTemplate(`
		resource "commercetools_shipping_zone" "standard" {
			name        = "the zone"
			description = "the description"

			location {
				country = "US"
				state = "Nevada"
			}
		}`, map[string]any{})
}

func testAccCheckShippingZoneDestroy(s *terraform.State) error {
	conn := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_shipping_zone" {
			continue
		}
		response, err := conn.Zones().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("shipping zone (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
