package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCategoryCreate_basic(t *testing.T) {
	resourceName := "commercetools_category.accessories"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCategoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_category.accessories_minimal", "name.en", "accessories_m"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "accessories"),
					resource.TestCheckResourceAttr(resourceName, "key", "accessories"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "Standard description"),
					resource.TestCheckResourceAttrPair(
						resourceName, "parent", "commercetools_category.accessories_base", "id"),
					resource.TestCheckResourceAttr(resourceName, "slug.en", "accessories"),
					resource.TestCheckResourceAttr(resourceName, "order_hint", "0.000016143365484621617765232"),
					resource.TestCheckResourceAttr(resourceName, "external_id", "some external id"),
					resource.TestCheckResourceAttr(resourceName, "meta_title.en", "meta text"),
					resource.TestCheckResourceAttr(resourceName, "meta_description.en", "meta description"),
					resource.TestCheckResourceAttr(resourceName, "meta_keywords.en", "keywords"),
					resource.TestCheckResourceAttr(resourceName, "assets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "assets.0.name.en", "My Product Video"),
					resource.TestCheckResourceAttr(resourceName, "assets.0.description.en", "Description"),
					resource.TestCheckResourceAttr(resourceName, "assets.0.sources.0.key", "image"),
				),
			},
			{
				Config: testAccCategoryUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "accessories"),
					resource.TestCheckResourceAttr(resourceName, "key", "accessories"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "Updated description"),
					resource.TestCheckResourceAttrPair(resourceName, "parent", "commercetools_category.accessories_base", "id"),
					resource.TestCheckResourceAttr(resourceName, "slug.en", "accessories_updated"),
					resource.TestCheckResourceAttr(resourceName, "order_hint", "0.000016143365484621617765232"),
					resource.TestCheckResourceAttr(resourceName, "external_id", "some external id"),
					resource.TestCheckResourceAttr(resourceName, "meta_title.en", "updated meta text"),
					resource.TestCheckResourceAttr(resourceName, "meta_description.en", "updated meta description"),
					resource.TestCheckResourceAttr(resourceName, "meta_keywords.en", "keywords, updated"),
					resource.TestCheckResourceAttr(resourceName, "assets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "assets.0.name.en", "Updated name"),
					resource.TestCheckResourceAttr(resourceName, "assets.0.description.en", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "assets.0.sources.0.key", "image"),
				),
			},
			{
				Config: testAccCategoryRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "accessories"),
					resource.TestCheckResourceAttr(resourceName, "key", "accessories_2"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "slug.en", "accessories_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "parent", "commercetools_category.accessories_base", "id"),
					resource.TestCheckResourceAttr(resourceName, "order_hint", "0.000016143365484621617765232"),
					resource.TestCheckResourceAttr(resourceName, "external_id", "some external id"),
					resource.TestCheckResourceAttr(resourceName, "meta_title.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "meta_description.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "meta_keywords.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "assets.#", "0"),
				),
			},
		},
	})
}

func TestAccCategoryRecreateAfterDelete(t *testing.T) {
	resourceName := "commercetools_category.accessories"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCategoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "accessories"),
				),
			},
			{
				Config: testAccCategoryConfig(),
				PreConfig: func() {
					client := getClient(testAccProvider.Meta())
					_, err := client.Categories().WithKey("accessories").Delete().Execute(context.Background())
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "accessories"),
				),
			},
		},
	})
}

func testAccCategoryConfig() string {
	return hclTemplate(`
		resource "commercetools_category" "accessories_base" {
			key = "accessories_b"

			name = {
				en = "accessories_b"
			}
			description = {
				en = "Standard description"
			}
			slug = {
				en = "accessories_b"
			}
			order_hint = "0.00001614336548703960465522"
		}

		resource "commercetools_category" "accessories_minimal" {
			name = {
				en = "accessories_m"
			}
			slug = {
				en = "accessories_m"
			}
		}

		resource "commercetools_category" "accessories" {
			key = "accessories"

			name = {
				en = "accessories"
			}
			description = {
				en = "Standard description"
			}
			slug = {
				en = "accessories"
			}

			parent 		= commercetools_category.accessories_base.id
			order_hint 	= "0.000016143365484621617765232"
			external_id = "some external id"

			meta_title = {
				en = "meta text"
			}
			meta_description = {
				en = "meta description"
			}
			meta_keywords = {
				en = "keywords"
			}
			assets {
				key = "some_key"
				name = {
					en = "My Product Video"
				}
				description = {
					en = "Description"
				}
				sources {
					uri = "https://www.w3.org/People/mimasa/test/imgformat/img/w3c_home.jpg"
					key = "image"
				}
			}
			depends_on = [commercetools_category.accessories_base]
		}
	`, map[string]any{})
}

func testAccCategoryUpdate() string {
	return hclTemplate(`
		resource "commercetools_category" "accessories_base" {
			key = "accessories_b"

			name = {
				en = "accessories_b"
			}
			description = {
				en = "Standard description"
			}
			slug = {
				en = "accessories_b"
			}
			order_hint = "0.00001614336548703960465522"
		}

		resource "commercetools_category" "accessories_minimal" {
			name = {
				en = "accessories_m"
			}
			slug = {
				en = "accessories_m"
			}
		}

		resource "commercetools_category" "accessories" {
			key = "accessories"

			name = {
				en = "accessories"
			}
			description = {
				en = "Updated description"
			}

			parent = commercetools_category.accessories_base.id
			order_hint = "0.000016143365484621617765232"
			external_id = "some external id"

			slug = {
				en = "accessories_updated"
			}
			meta_title = {
				en = "updated meta text"
			}
			meta_description = {
				en = "updated meta description"
			}
			meta_keywords = {
				en = "keywords, updated"
			}
			assets {
				key = "some_key"
				name = {
					en = "Updated name"
				}
				description = {
					en = "Updated description"
				}
				sources {
					uri = "https://www.w3.org/People/mimasa/test/imgformat/img/w3c_home.jpg"
					key = "image"

					dimensions {
						w = 10
						h = 20
					}

				}
			}
		}
	`, map[string]any{})
}

func testAccCategoryRemoveProperties() string {
	return hclTemplate(`
		resource "commercetools_category" "accessories_base" {
			key = "accessories_b"

			name = {
				en = "accessories_b"
			}
			description = {
				en = "Standard description"
			}
			slug = {
				en = "accessories_b"
			}
			order_hint = "0.00001614336548703960465522"
		}

		resource "commercetools_category" "accessories" {
			key = "accessories_2"

			name = {
				en = "accessories"
			}
			description = {
				en = "Updated description"
			}
			slug = {
				en = "accessories_updated"
			}
			parent 		= commercetools_category.accessories_base.id
			order_hint 	= "0.000016143365484621617765232"
			external_id = "some external id"
		}
	`, map[string]any{})
}

func testAccCategoryDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_category" {
			continue
		}
		response, err := client.Categories().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("category (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
