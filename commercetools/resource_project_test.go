package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
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
						"commercetools_project_settings.acctest_project_settings", "messages.0.enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.url", "https://example.com/oauth/token",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.authorization_header", "Bearer secret",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_input_type", "CartValue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.0.country_tax_rate_fallback_enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.0.delete_days_after_last_modification", "7"),

					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["commercetools_project_settings.acctest_project_settings"]
						if !ok {
							return fmt.Errorf("Project not found")
						}

						if rs.Primary.ID == "" {
							return fmt.Errorf("No Project ID found")
						}

						client := getClient(testAccProvider.Meta())
						result, err := client.Get().Execute(context.Background())
						if err != nil {
							return err
						}
						if result == nil {
							return fmt.Errorf("resource not found")
						}

						assert.True(t, *result.Carts.CountryTaxRateFallbackEnabled)
						assert.EqualValues(t, result.Messages.Enabled, true)
						assert.EqualValues(t, result.ExternalOAuth.Url, "https://example.com/oauth/token")
						assert.EqualValues(t, result.ExternalOAuth.AuthorizationHeader, "****")
						assert.EqualValues(t, result.Countries, []string{"NL", "DE", "US"})
						assert.EqualValues(t, result.Languages, []string{"nl", "de", "en", "en-US"})
						assert.EqualValues(t, result.Currencies, []string{"EUR", "USD"})
						assert.Equal(t, *result.Carts.DeleteDaysAfterLastModification, 7)
						assert.Equal(t, result.ShippingRateInputType, platform.CartValueType(platform.CartValueType{}))
						return nil
					},
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
						"commercetools_project_settings.acctest_project_settings", "messages.0.enabled", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.url", "https://new-example.com/oauth/token",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.authorization_header", "Bearer new-secret",
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
						"commercetools_project_settings.acctest_project_settings", "carts.0.country_tax_rate_fallback_enabled", "false",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.0.delete_days_after_last_modification", "21",
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
						"commercetools_project_settings.acctest_project_settings", "messages.0.enabled", "false",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.url",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.authorization_header",
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
						"commercetools_project_settings.acctest_project_settings", "carts.0.country_tax_rate_fallback_enabled",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.0.delete_days_after_last_modification",
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
						"commercetools_project_settings.acctest_project_settings", "messages.0.enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.url", "https://example.com/oauth/token",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "external_oauth.0.authorization_header", "Bearer secret",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "shipping_rate_input_type", "CartValue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_project_settings.acctest_project_settings", "carts.0.country_tax_rate_fallback_enabled", "true",
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

			external_oauth {
				url = "https://example.com/oauth/token"
				authorization_header = "Bearer secret"
			}

			messages {
			  enabled = true
			}

			carts {
              country_tax_rate_fallback_enabled = true
              delete_days_after_last_modification = 7
            }

			shipping_rate_input_type = "CartValue"
		}`
}

func testAccProjectConfigUpdate() string {
	return `
		resource "commercetools_project_settings" "acctest_project_settings" {
			name       = "Test this thing new"
			countries  = ["nL", "De", "us", "gb"]
			currencies = ["Eur", "UsD", "GbP"]
			languages  = ["NL", "dE", "en", "eN-uS", "Fr"]
			external_oauth {
				url = "https://new-example.com/oauth/token"
				authorization_header = "Bearer new-secret"
			}
			messages {
			  enabled = false
			}

			carts {
              country_tax_rate_fallback_enabled = false
              delete_days_after_last_modification = 21
            }

			enable_search_index_products = true
			enable_search_index_orders = true

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
			messages {
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
