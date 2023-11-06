package product_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestProductResource_Create(t *testing.T) {
	resourceName := "commercetools_product.test_product"
	name := "test_product"
	key := "test-product-key"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfigBasic(name, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					// func(s *terraform.State) error {
					// 	fmt.Println("Sleeping...")
					// 	time.Sleep(10 * time.Second)
					// 	return nil
					// },
				),
			},
		},
	})
}

func testProductDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product" {
			continue
		}
		response, err := client.Products().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testAccProductConfigBasic(identifier, key string) string {
	return utils.HCLTemplate(`
		resource "commercetools_product_type" "default_product_type" {
			key         = "default-product-type"
			name        = "Default Product Type"
			description = "Default Product Type Description"
		
			attribute {
				constraint = "None"
				input_hint = "SingleLine"
				input_tip = {
					en-GB = "SKU name"
				}
				label = {
					en-GB = "Name"
				}
				name       = "name"
				required   = false
				searchable = false
				type {
					name = "ltext"
				}
			}
		
			attribute {
				constraint = "None"
				input_hint = "SingleLine"
				input_tip = {
					en-GB = "SKU Description"
				}
				label = {
					en-GB = "Description"
				}
				name       = "description"
				required   = false
				searchable = false
				type {
					name = "ltext"
				}
			}
		}
		
		resource "commercetools_category" "category1" {
			key = "category-1-key"
		
			name = {
				en-GB = "Category One"
			}
			description = {
				en-GB = "Category One description"
			}
			slug = {
				en-GB = "category_1"
			}
			meta_title = {
				en-GB = "Category One Meta Title"
			}
		}
		
		resource "commercetools_tax_category" "external_shipping_tax" {
			key         = "external-shipping-tax"
			name        = "External Shipping Tax"
			description = "External Shipping Tax Description"
		}
		
		resource "commercetools_state" "product_state_for_sale" {
			key  = "product-for-sale"
			type = "ProductState"
			name = {
			en-GB = "For Sale"
			}
			description = {
			en-GB = "Regularly stocked product."
			}
			initial = true
		}
		
		resource "commercetools_product" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = {
				en-GB = "test-product-name"
			}
			slug = {
				en-GB = "test-product-slug"
			}
			description = {
				en-GB = "Test product description"
			}
			product_type_id = commercetools_product_type.default_product_type.id
			publish         = false
			categories = [
				commercetools_category.category1.id,
			]
			tax_category_id = commercetools_tax_category.external_shipping_tax.id
			meta_title = {
				en-GB = "meta-title"
			}
			meta_description = {
				en-GB = "meta-description"
			}
			meta_keywords = {
				en-GB = "meta-keywords"
			}
			state_id = commercetools_state.product_state_for_sale.id
			master_variant {
				key = "master-variant-key"
				sku = "100000"
				attribute {
					name  = "name"
					value = jsonencode({ "en-GB" : "Test product basic variant" })
				}
				attribute {
					name  = "description"
					value = jsonencode({ "en-GB" : "Test product basic variant description" })
				}
				price {
					key = "base_price_eur"
					value {
						cent_amount   = 1000000
						currency_code = "EUR"
					}
				}
				price {
					key = "base_price_gbr"
					value {
						cent_amount   = 872795
						currency_code = "GBP"
					}
				}
			}
		
			variant {
				key = "variant-1-key"
				sku = "100001"
				attribute {
					name  = "name"
					value = jsonencode({ "en-GB" : "Test product variant one" })
				}
				attribute {
					name  = "description"
					value = jsonencode({ "en-GB" : "Test product variant one description" })
				}
				price {
					key = "base_price_eur"
					value {
						cent_amount   = 1010000
						currency_code = "EUR"
					}
				}
				price {
					key = "base_price_gbr"
					value {
						cent_amount   = 880299
						currency_code = "GBP"
					}
				}
			}
		}`,
		map[string]any{
			"identifier": identifier,
			"key":        key,
		})
}
