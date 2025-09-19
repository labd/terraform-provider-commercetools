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
					resource.TestCheckResourceAttr(resourceName, "messages.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "carts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "shopping_lists.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_oauth.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_orders", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_customers", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_products", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_product_search", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_business_units", "false"),
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

						assert.EqualValues(t, []string{"NL", "DE", "US"}, result.Countries)
						assert.EqualValues(t, []string{"nl", "de", "en", "en-US"}, result.Languages)
						assert.EqualValues(t, []string{"EUR", "USD"}, result.Currencies)
						assert.EqualValues(t, false, result.Messages.Enabled)
						assert.EqualValues(t, utils.IntRef(15), result.Messages.DeleteDaysAfterCreation)
						assert.Equal(t, 90, *result.Carts.DeleteDaysAfterLastModification)
						assert.Equal(t, false, *result.Carts.CountryTaxRateFallbackEnabled)
						assert.Equal(t, utils.Ref(platform.RoundingModeHalfEven), result.Carts.PriceRoundingMode)
						assert.Equal(t, utils.Ref(platform.RoundingModeHalfEven), result.Carts.TaxRoundingMode)
						assert.Equal(t, 90, *result.ShoppingLists.DeleteDaysAfterLastModification)
						assert.Nil(t, result.ShippingRateInputType)
						assert.Nil(t, result.BusinessUnits)
						return nil
					},
				),
			},
			{
				Config: testAccProjectConfigUpdate("acctest_project_settings"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test this thing new"),
					resource.TestCheckResourceAttr(resourceName, "countries.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "currencies.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "languages.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "messages.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "messages.0.delete_days_after_creation", "15"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_orders", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_customers", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_products", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_product_search", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_search_index_business_units", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "external_oauth.0.url", "https://new-example.com/oauth/token"),
					resource.TestCheckResourceAttr(
						resourceName, "external_oauth.0.authorization_header", "Bearer new-secret"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_input_type", "CartClassification"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.0.key", "Small"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.0.label.en", "Small"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.0.label.nl", "Klein"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.1.key", "Medium"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.1.label.en", "Medium"),
					resource.TestCheckResourceAttr(
						resourceName, "shipping_rate_cart_classification_value.1.label.nl", "Middel"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.0.country_tax_rate_fallback_enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.0.delete_days_after_last_modification", "21"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.0.price_rounding_mode", "HalfUp"),
					resource.TestCheckResourceAttr(
						resourceName, "carts.0.tax_rounding_mode", "HalfUp"),
					resource.TestCheckResourceAttr(
						resourceName, "shopping_lists.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "shopping_lists.0.delete_days_after_last_modification", "14"),
					resource.TestCheckResourceAttr(resourceName, "business_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "business_units.0.my_business_unit_status_on_creation", "Inactive"),
					resource.TestCheckResourceAttr(resourceName, "business_units.0.my_business_unit_associate_role_key_on_creation", "my-role"),
				),
			},
			{
				Config: testAccProjectConfigDeleteOAuthAndCarts("acctest_project_settings"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Test this thing new"),
					resource.TestCheckResourceAttr(resourceName, "countries.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "currencies.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "languages.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "messages.0.enabled", "false"),
					resource.TestCheckNoResourceAttr(resourceName,
						"external_oauth.0.url"),
					resource.TestCheckNoResourceAttr(resourceName,
						"external_oauth.0.authorization_header"),
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
						resourceName, "carts.0.country_tax_rate_fallback_enabled"),
					resource.TestCheckNoResourceAttr(
						resourceName, "carts.0.delete_days_after_last_modification"),
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
		}`, map[string]any{
		"identifier": identifier,
	})
}

func testAccProjectConfigUpdate(identifier string) string {
	return utils.HCLTemplate(`
		resource commercetools_associate_role "my-role" {
		  key = "my-role"
		  name = "My role"
		  permissions = [
			"UpdateAssociates"
		  ]
		}
	
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
				enabled = true
				delete_days_after_creation = 15
			}

			carts {
				country_tax_rate_fallback_enabled = true
				delete_days_after_last_modification = 21
				price_rounding_mode = "HalfUp"
				tax_rounding_mode = "HalfUp"
			}
	
			shopping_lists {
				delete_days_after_last_modification = 14
			}

		   	enable_search_index_orders         = true
	       	enable_search_index_customers      = true
   			enable_search_index_products       = true
   			enable_search_index_product_search = true
   			enable_search_index_business_units = true

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
	
			business_units {
			  my_business_unit_status_on_creation             = "Inactive"
              my_business_unit_associate_role_key_on_creation = commercetools_associate_role.my-role.key
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
