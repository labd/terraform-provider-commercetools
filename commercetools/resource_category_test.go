package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccCategoryCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCategoryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "name.en", "accessories",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "key", "accessories",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttrPair(
						"commercetools_category.accessories", "parent", "commercetools_category.accessories_base", "id",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "slug.en", "accessories",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "order_hint", "0.000016143365484621617765232",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "meta_title.en", "meta text",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "meta_description.en", "meta description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "meta_keywords.en", "keywords",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "assets.0.name.en", "My Product Video",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "assets.0.description.en", "Description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "assets.0.sources.0.key", "image",
					),
				),
			},
			{
				Config: testAccCategoryUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "name.en", "accessories",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "key", "accessories",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "description.en", "Updated description",
					),
					resource.TestCheckResourceAttrPair(
						"commercetools_category.accessories", "parent", "commercetools_category.accessories_base", "id",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "slug.en", "accessories_updated",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "order_hint", "0.000016143365484621617765232",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "meta_title.en", "updated meta text",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "meta_description.en", "updated meta description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "meta_keywords.en", "keywords, updated",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "assets.0.name.en", "Updated name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "assets.0.description.en", "Updated description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "assets.0.sources.0.key", "image",
					),
				),
			},
			{
				Config: testAccCategoryRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "name.en", "accessories",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "key", "accessories_2",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "description.en", "Updated description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "slug.en", "accessories_updated",
					),
					resource.TestCheckResourceAttrPair(
						"commercetools_category.accessories", "parent", "commercetools_category.accessories_base", "id",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "order_hint", "0.000016143365484621617765232",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_category.accessories", "meta_title",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_category.accessories", "meta_description",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_category.accessories", "meta_keywords",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_category.accessories", "assets",
					),
				),
			},
		},
	})
}

func testAccCategoryConfig() string {
	return `
	resource "commercetools_category" "accessories_base" {
		name = {
			en = "accessories_b"
		}
		key = "accessories_b"
		description = {
			en = "Standard description"
		}
		slug = {
			en = "accessories_b"
		}
		order_hint = "0.00001614336548703960465522"
	}

	resource "commercetools_category" "accessories" {
		name = {
			en = "accessories"
		}
		key = "accessories"
		description = {
			en = "Standard description"
		}
		parent = "${commercetools_category.accessories_base.id}"
		slug = {
			en = "accessories"
		}
		order_hint = "0.000016143365484621617765232"
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
	}  `
}

func testAccCategoryUpdate() string {
	return `
	resource "commercetools_category" "accessories_base" {
		name = {
			en = "accessories_b"
		}
		key = "accessories_b"
		description = {
			en = "Standard description"
		}
		slug = {
			en = "accessories_b"
		}
		order_hint = "0.00001614336548703960465522"
	}

	resource "commercetools_category" "accessories" {
		name = {
			en = "accessories"
		}
		key = "accessories"
		description = {
			en = "Updated description"
		}
		parent = "${commercetools_category.accessories_base.id}"
		slug = {
			en = "accessories_updated"
		}
		order_hint = "0.000016143365484621617765232"
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
			}
		}
	}  `
}

func testAccCategoryRemoveProperties() string {
	return `
	resource "commercetools_category" "accessories_base" {
		name = {
			en = "accessories_b"
		}
		key = "accessories_b"
		description = {
			en = "Standard description"
		}
		slug = {
			en = "accessories_b"
		}
		order_hint = "0.00001614336548703960465522"
	}

	resource "commercetools_category" "accessories" {
		name = {
			en = "accessories"
		}
		key = "accessories_2"
		description = {
			en = "Updated description"
		}
		parent = "${commercetools_category.accessories_base.id}"
		slug = {
			en = "accessories_updated"
		}
		order_hint = "0.000016143365484621617765232"
	}  `
}
