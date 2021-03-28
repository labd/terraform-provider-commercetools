package commercetools

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestCategoryCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testCategory(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "name.en", "accessories"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "key", "bananas123"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "description.en", "da"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "slug.en", "bananas_accessories"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "order_hint", "0.001"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "external_id", "iddqd"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_title.en", "foo"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_description.en", "bar"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_keywords.en", "baz"),
				),
			},
			{
				Config: testCategoryUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "name.en", "accessories"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "key", "bananas123"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "description.en", "vi very viniversum vivus vicy"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "slug.en", "bananas_accessories"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "order_hint", "0.002"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "external_id", "idclip"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_title.en", "baz"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_description.en", "foo"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_keywords.en", "bar"),
				),
			},{
				Config: testCreateChildCategory(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_category.bracelets", "name.en", "bracelets"),
					resource.TestCheckResourceAttr("commercetools_category.bracelets", "key", "bracelets123"),
					resource.TestCheckResourceAttr("commercetools_category.bracelets", "description.en", "nice bracelets"),
					resource.TestCheckResourceAttr("commercetools_category.bracelets", "slug.en", "foo_bracelets"),
					resource.TestCheckResourceAttr("commercetools_category.bracelets", "parent_key", "bananas123"),

				),
			},
			{
				Config: testChangeParents(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_category.bracelets", "parent_key", "new_parent"),
				),
			},
		},
	})
}

func TestCategoryCreate_withAssets(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAddAssets(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "key", "accessoriesz_with_assets"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.name.en", "ass1"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.key", "it_a_nice"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.sources.0.uri", "http://google.com"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.sources.0.key", "keywest"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.sources.0.dimensions.w", "240"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.sources.0.dimensions.h", "240"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.sources.0.content_type", "png"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.sources.1.uri", "http://tapuz.com"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.description.en", "terraform is so much fun"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.tags.0", "banana"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.0.tags.1", "tapuz"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.1.name.en", "ass2"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz_with_assets", "assets.1.sources.0.uri", "http://nice.com"),
				),
			},
		},
	})
}



func testCategory() string {
	return `resource "commercetools_category" "accessoriesz" {
			name = {
				en = "accessories"
			}
			key = "bananas123"
			description = {
				en = "da"
			}
			slug = {
				en = "bananas_accessories"
			}
			order_hint = "0.001"
			external_id = "iddqd"
			meta_title = { en = "foo" }
			meta_description = { en = "bar" }
			meta_keywords = { en = "baz" }
		}`
}

func testCategoryUpdate() string {
	return `resource "commercetools_category" "accessoriesz" {
			name = {
				en = "accessories"
			}
			key = "bananas123"
			description = {
				en = "vi very viniversum vivus vicy"
			}
			slug = {
				en = "bananas_accessories"
			}
			order_hint = "0.002"
			external_id = "idclip"
			meta_title = { en = "baz" }
			meta_description = { en = "foo" }
			meta_keywords = { en = "bar" }
		}`
}

func testCreateChildCategory() string {
	return fmt.Sprintf(`%s 

resource "commercetools_category" "bracelets" {
			name = {
				en = "bracelets"
			}
			key = "bracelets123"
			description = {
				en = "nice bracelets"
			}
			slug = {
				en = "foo_bracelets"
			}
			order_hint = "0.008"
			parent_key = "bananas123"
		}`,testCategoryUpdate())
}





func testChangeParents() string {
	return `

resource "commercetools_category" "accessoriesz" {
			name = {
				en = "accessories"
			}
			key = "bananas123"
			description = {
				en = "vi very viniversum vivus vicy"
			}
			slug = {
				en = "bananas_accessories"
			}
			order_hint = "0.002"
			external_id = "idclip"
			meta_title = { en = "baz" }
			meta_description = { en = "foo" }
			meta_keywords = { en = "bar" }
}


resource "commercetools_category" "accessoriesxx" {
			name = {
				en = "accessoriesxx"
			}
			key = "new_parent"
			description = {
				en = "fooo"
			}
			slug = {
				en = "bar_accessories"
			}
			order_hint = "0.002"
}

resource "commercetools_category" "bracelets" {
			name = {
				en = "bracelets"
			}
			key = "bracelets123"
			description = {
				en = "nice bracelets"
			}
			slug = {
				en = "foo_bracelets"
			}
			order_hint = "0.008"
			parent_key = "new_parent"
}
`
}

func testAddAssets() string {
	return `

resource "commercetools_category" "accessoriesz_with_assets" {
			name = {
				en = "banana"
			}
			key = "accessoriesz_with_assets"
			description = {
				en = "accessoriesz_with_assets"
			}
			slug = {
				en = "accessoriesz_with_assets"
			}
			order_hint = "0.002"
			assets {
				name = {
					en = "ass1"
				}
				key = "it_a_nice"
				sources {
					uri = "http://google.com"
					key = "keywest"
					content_type = "png"
					dimensions = {
						w = 240
						h = 240
					}
				}
				sources {
					uri = "http://tapuz.com"
				}
				description = {
					en = "terraform is so much fun"
				}
				tags = ["banana","tapuz"]
			}
			assets {
				name = {
					en = "ass2"
				}
				sources {
					uri = "http://nice.com"
				}
			}
}

`
}

