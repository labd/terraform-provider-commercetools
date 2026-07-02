package custom_product_type_test

import (
	"context"
	"fmt"
	"testing"

	acctest "github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccProductType_DataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductTypeDestroy,
		Steps: []resource.TestStep{{
			Config: testAccConfigCreateProductType(),
			Check: resource.ComposeTestCheckFunc(
				func(s *terraform.State) error {
					client, err := acctest.GetClient()
					if err != nil {
						return err
					}
					result, err := client.ProductTypes().WithKey("test").Get().Execute(context.Background())
					if err != nil {
						return err
					}
					assert.NotNil(t, result)
					assert.Equal(t, "test", result.Key)
					assert.Equal(t, "my-attribute", result.Attributes[0].Name)
					return nil
				},
			),
		},
			{
				Config: testAccConfigWithDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.commercetools_product_type.test_product_type", "key", "test"),
					func(s *terraform.State) error {
						client, err := acctest.GetClient()
						if err != nil {
							return err
						}
						result_product_type, err := client.ProductTypes().WithKey("test").Get().Execute(context.Background())
						if err != nil {
							return err
						}
						assert.NotNil(t, result_product_type)

						// Verify the data source returns the correct ID
						rs, ok := s.RootModule().Resources["data.commercetools_product_type.test_product_type"]
						if !ok {
							return fmt.Errorf("data source not found")
						}
						assert.Equal(t, result_product_type.ID, rs.Primary.ID)
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckProductTypeDestroy(s *terraform.State) error {
	return nil
}

func testAccConfigCreateProductType() string {
	return utils.HCLTemplate(`
	resource "commercetools_product_type" "test"  {
		key = "test"
		name = "Test Product Type"
		description = "Test product type for data source testing"

		attribute {
			name = "my-attribute"
			label = {
				en = "My Attribute"
			}
			type {
				name = "text"
			}
		}
	}
	`, map[string]any{})
}

func testAccConfigWithDataSource() string {
	return utils.HCLTemplate(`
		resource "commercetools_product_type" "test"  {
			key = "test"
			name = "Test Product Type"
			description = "Test product type for data source testing"

			attribute {
				name = "my-attribute"
				label = {
					en = "My Attribute"
				}
				type {
					name = "text"
				}
			}
		}

		data "commercetools_product_type" "test_product_type"{
			key = "test"
		}

	`, map[string]any{})
}
