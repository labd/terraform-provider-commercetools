package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccShippingMethodStores(t *testing.T) {
	identifier := "stores"
	resourceName := "commercetools_shipping_method.stores"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodWithoutStores(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "0"),
				),
			},
			{
				Config: testAccShippingMethodWithStores(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "stores.*", "my-store"),
				),
			},
			{
				Config: testAccShippingMethodWithoutStores(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "0"),
				),
			},
		},
	})
}

func testAccShippingMethodWithoutStores(identifier string) string {
	return hclTemplate(`
		resource "commercetools_tax_category" "{{ .identifier }}" {
			name = "test"
			key  = "test"
		}

		resource "commercetools_shipping_method" "{{ .identifier }}" {
			name            = "shipping method with stores"
			tax_category_id = commercetools_tax_category.{{ .identifier }}.id
		}
	`, map[string]any{
		"identifier": identifier,
	})
}

func testAccShippingMethodWithStores(identifier string) string {
	return hclTemplate(`
		resource "commercetools_tax_category" "{{ .identifier }}" {
			name = "test"
			key  = "test"
		}

		resource "commercetools_store" "my-store-{{ .identifier }}" {
			key = "my-store"
			name = {
				en-US = "My store"
			}
			countries = ["NL", "BE"]
			languages = ["nl-NL"]
		}

		resource "commercetools_shipping_method" "{{ .identifier }}" {
			name            = "shipping method with stores"
			tax_category_id = commercetools_tax_category.{{ .identifier }}.id
			stores          = [commercetools_store.my-store-{{ .identifier }}.key]
		}
	`, map[string]any{
		"identifier": identifier,
	})
}
