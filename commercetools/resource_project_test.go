package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccProjectCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "name", "Test this thing",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "countries.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "currencies.#", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "languages.#", "4",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "messages.enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.url", "https://example.com/oauth/token",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.authorization_header", "Bearer secret",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_input_type", "CartValue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.country_tax_rate_fallback_enabled", "true",
					),
				),
			},
			{
				Config: testAccProjectConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "name", "Test this thing new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "countries.#", "4",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "currencies.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "languages.#", "5",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "messages.enabled", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.url", "https://new-example.com/oauth/token",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.authorization_header", "Bearer new-secret",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_input_type", "CartClassification",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.#", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.0.key", "Small",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.0.label.en", "Small",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.0.label.nl", "Klein",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.1.key", "Medium",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.1.label.en", "Medium",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.1.label.nl", "Middel",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.country_tax_rate_fallback_enabled", "false",
					),
				),
			},
			{
				Config: testAccProjectConfigDeleteOAuthAndCarts(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "name", "Test this thing new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "countries.#", "4",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "currencies.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "languages.#", "5",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "messages.enabled", "false",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.url",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.authorization_header",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_input_type", "CartClassification",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.#", "1",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.0.key", "Small",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.0.label.en", "Small",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_cart_classification_value.0.label.nl", "Klein",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.country_tax_rate_fallback_enabled",
					),
				),
			},
			// Running this step again so project settings match what later shipping_zone_rate_test will need
			{
				Config: testAccProjectConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "name", "Test this thing",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "countries.#", "3",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "currencies.#", "2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "languages.#", "4",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "messages.enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.url", "https://example.com/oauth/token",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.authorization_header", "Bearer secret",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_input_type", "CartValue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.country_tax_rate_fallback_enabled", "true",
					),
				),
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	return nil
}

func testAccProjectConfig() string {
	return `
		resource "commercetools_project_settings" "acctest_project_settings" {
			name       = "Test this thing"
			countries  = ["NL", "DE", "US"]
			currencies = ["EUR", "USD"]
			languages  = ["nl", "de", "en", "en-US"]
			external_oauth = {
				url = "https://example.com/oauth/token"
				authorization_header = "Bearer secret"
			}
			messages = {
			  enabled = true
			}

			carts = {
              country_tax_rate_fallback_enabled = true
            }

			shipping_rate_input_type = "CartValue"
                        

		}`
}

func testAccProjectConfigUpdate() string {
	return `
		resource "commercetools_project_settings" "acctest_project_settings" {
			name       = "Test this thing new"
			countries  = ["NL", "DE", "US", "GB"]
			currencies = ["EUR", "USD", "GBP"]
			languages  = ["nl", "de", "en", "en-US", "fr"]
			external_oauth = {
				url = "https://new-example.com/oauth/token"
				authorization_header = "Bearer new-secret"
			}
			messages = {
			  enabled = false
			}

			carts = {
              country_tax_rate_fallback_enabled = false
            }

            shipping_rate_input_type = "CartClassification"
			shipping_rate_cart_classification_value {
				key = "Small"
				label = {
					"en" = "Small"
					"nl" = "Klein"
				}
			}

			shipping_rate_cart_classification_value {
				key = "Medium"
				label = {
					"en" = "Medium"
					"nl" = "Middel"
				}
			}
		}`
}

func testAccProjectConfigDeleteOAuthAndCarts() string {
	return `
		resource "commercetools_project_settings" "acctest_project_settings" {
			name       = "Test this thing new"
			countries  = ["NL", "DE", "US", "GB"]
			currencies = ["EUR", "USD", "GBP"]
			languages  = ["nl", "de", "en", "en-US", "fr"]
			messages = {
			  enabled = false
			}

			shipping_rate_input_type = "CartClassification"
			shipping_rate_cart_classification_value {
				key = "Small"
				label = {
					"en" = "Small"
					"nl" = "Klein"
				}
			}
		}`
}
