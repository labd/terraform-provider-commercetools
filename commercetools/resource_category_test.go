package commercetools

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

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
			external_id = "idclip"
			meta_title = { en = "baz" }
			meta_description = { en = "foo" }
			meta_keywords = { en = "bar" }
		}`
}


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
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "external_id", "idclip"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_title.en", "baz"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_description.en", "foo"),
					resource.TestCheckResourceAttr("commercetools_category.accessoriesz", "meta_keywords.en", "bar"),
				),
			},
		},
	})
}
