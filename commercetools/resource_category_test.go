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
						"commercetools_category.accessories", "description.en", "Standard description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "slug.en", "accessories",
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
						"commercetools_category.accessories", "description.en", "Updated description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "slug.en", "accessories_updated",
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
						"commercetools_category.accessories", "description.en", "Updated description",
					),
					resource.TestCheckResourceAttr(
						"commercetools_category.accessories", "slug.en", "accessories_updated",
					),
					resource.TestCheckNoResourceAttr(
						"commercetools_category.accessories", "meta_title",
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
			en = "meta text"
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
		key = "accessories"
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
