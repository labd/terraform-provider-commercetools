package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccShippingMethod_storesAndCarrier(t *testing.T) {
	resourceName := "commercetools_shipping_method.stores"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodStoresConfig(false, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "carrier", ""),
				),
			},
			{
				Config: testAccShippingMethodStoresConfig(true, "DHL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "stores.*", "shipping-method-store"),
					resource.TestCheckResourceAttr(resourceName, "carrier", "DHL"),
				),
			},
			{
				Config: testAccShippingMethodStoresConfig(true, "UPS"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(resourceName, "stores.*", "shipping-method-store"),
					resource.TestCheckResourceAttr(resourceName, "carrier", "UPS"),
				),
			},
			{
				Config: testAccShippingMethodStoresConfig(false, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "stores.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "carrier", ""),
				),
			},
		},
	})
}

func testAccShippingMethodStoresConfig(withStores bool, carrier string) string {
	return hclTemplate(`
		resource "commercetools_tax_category" "stores" {
			name = "shipping-method-stores"
			key  = "shipping-method-stores"
		}

		resource "commercetools_store" "stores" {
			key = "shipping-method-store"
			name = {
				en = "Shipping method store"
			}
		}

		resource "commercetools_shipping_method" "stores" {
			name            = "shipping-method-stores"
			key             = "shipping-method-stores"
			tax_category_id = commercetools_tax_category.stores.id
			{{ if .withStores }}
			stores          = [commercetools_store.stores.key]
			{{ end }}
			{{ if .carrier }}
			carrier         = "{{ .carrier }}"
			{{ end }}
		}
	`, map[string]any{
		"withStores": withStores,
		"carrier":    carrier,
	})
}
