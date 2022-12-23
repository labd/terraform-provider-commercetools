package project_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccProjectCreate_basic(t *testing.T) {
	resourceName := "commercetools_project_settings.acctest_project_settings"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig("acctest_project_settings"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test this thing"),
					resource.TestCheckResourceAttr(resourceName, "countries.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "currencies.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "languages.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "messages.enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "external_oauth.url", "https://example.com/oauth/token"),
					resource.TestCheckResourceAttr(
						resourceName, "external_oauth.authorization_header", "Bearer secret"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_input_type", "CartValue"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.country_tax_rate_fallback_enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.delete_days_after_last_modification", "7"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Project not found")
						}

						if rs.Primary.ID == "" {
							return fmt.Errorf("No Project ID found")
						}

						client, err := acctest.GetClient()
						if err != nil {
							return err
						}
						result, err := client.Get().Execute(context.Background())
						if err != nil {
							return err
						}
						if result == nil {
							return fmt.Errorf("resource not found")
						}

						assert.True(t, *result.Carts.CountryTaxRateFallbackEnabled)
						assert.EqualValues(t, result.Messages.Enabled, true)
						assert.EqualValues(t, result.Messages.DeleteDaysAfterCreation, utils.IntRef(90))
						assert.EqualValues(t, result.ExternalOAuth.Url, "https://example.com/oauth/token")
						assert.EqualValues(t, result.ExternalOAuth.AuthorizationHeader, "****")
						assert.EqualValues(t, result.Countries, []string{"NL", "DE", "US"})
						assert.EqualValues(t, result.Languages, []string{"nl", "de", "en", "en-US"})
						assert.EqualValues(t, result.Currencies, []string{"EUR", "USD"})
						assert.Equal(t, 7, *result.Carts.DeleteDaysAfterLastModification)
						assert.Equal(t, platform.CartValueType(platform.CartValueType{}), result.ShippingRateInputType)
						return nil
					},
				),
			},
			// {
			// 	Config: testAccProjectConfigUpdate("acctest_project_settings"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr(resourceName, "name", "Test this thing new"),
			// 		resource.TestCheckResourceAttr(resourceName, "countries.#", "4"),
			// 		resource.TestCheckResourceAttr(resourceName, "currencies.#", "3"),
			// 		resource.TestCheckResourceAttr(resourceName, "languages.#", "5"),
			// 		resource.TestCheckResourceAttr(resourceName, "messages.enabled", "false"),
			// 		resource.TestCheckResourceAttr(resourceName, "messages.delete_days_after_creation", "15"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "external_oauth.url", "https://new-example.com/oauth/token"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "external_oauth.authorization_header", "Bearer new-secret"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_input_type", "CartClassification"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.#", "2"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.0.key", "Small"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.0.label.en", "Small"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.0.label.nl", "Klein"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.1.key", "Medium"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.1.label.en", "Medium"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "shipping_rate_cart_classification_value.1.label.nl", "Middel"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "carts.country_tax_rate_fallback_enabled", "false"),
			// 		resource.TestCheckResourceAttr(
			// 			resourceName, "carts.delete_days_after_last_modification", "21"),
			// 	),
			// },
			{
				Config: testAccProjectConfigDeleteOAuthAndCarts("acctest_project_settings"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test this thing new"),
					resource.TestCheckResourceAttr(resourceName, "countries.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "currencies.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "languages.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "messages.enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "external_oauth.url"),
					resource.TestCheckNoResourceAttr(resourceName, "external_oauth.authorization_header"),
					resource.TestCheckResourceAttr(resourceName,
						"shipping_rate_input_type", "CartClassification"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.0.key", "Small"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.0.label.en", "Small"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.0.label.nl", "Klein"),
					resource.TestCheckNoResourceAttr(
						resourceName, "carts.country_tax_rate_fallback_enabled"),
					resource.TestCheckNoResourceAttr(
						resourceName, "carts.delete_days_after_last_modification"),
				),
			},
			// Running this step again so project settings match what later shipping_zone_rate_test will need
			{
				Config: testAccProjectConfig("acctest_project_settings"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test this thing"),
					resource.TestCheckResourceAttr(resourceName, "countries.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "currencies.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "languages.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "messages.enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "external_oauth.url", "https://example.com/oauth/token"),
					resource.TestCheckResourceAttr(
						resourceName, "external_oauth.authorization_header", "Bearer secret"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_input_type", "CartValue"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.country_tax_rate_fallback_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	return nil
}

func testAccProjectConfig(identifier string) string {
	return utils.HCLTemplate(`
		resource "commercetools_project_settings" "{{ .identifier }}" {
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
				delete_days_after_creation = 90
			}

			carts {
				country_tax_rate_fallback_enabled = true
				delete_days_after_last_modification = 7
			}

			shipping_rate_input_type = "CartValue"
		}`, map[string]any{
		"identifier": identifier,
	})
}

func testAccProjectConfigUpdate(identifier string) string {
	return utils.HCLTemplate(`
		resource "commercetools_project_settings" "{{ .identifier }}" {
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
				delete_days_after_creation = 15
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
		}`, map[string]any{
		"identifier": identifier,
	})
}

func testAccProjectConfigDeleteOAuthAndCarts(identifier string) string {
	return utils.HCLTemplate(`
		resource "commercetools_project_settings" "{{ .identifier }}" {
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
		}`, map[string]any{
		"identifier": identifier,
	})
}
