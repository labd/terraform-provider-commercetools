package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccCustomObjectCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomObjectNumber(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "container", "foobar",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "key", "value",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "value", "10",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "version", "1",
					),
				),
			},
			{
				Config: testAccCustomObjectNumberUpdateValue(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "container", "foobar",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "key", "value",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "value", "20",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "version", "2",
					),
				),
			},
			{
				Config: testAccCustomObjectNumberUpdateKey(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "container", "foobar",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "key", "newvalue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "value", "20",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "version", "1",
					),
				),
			},
			{
				Config: testAccCustomObjectNumberUpdateContainer(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "container", "newbar",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "key", "newvalue",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "value", "20",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_number", "version", "1",
					),
				),
			},
		},
	})
}

func TestAccCustomObjectCreate_object(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomObjectNestedData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_nested", "container", "foobar",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_nested", "key", "nested",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_nested", "value", "{\"address\":{\"number\":10,\"street\":\"foo\"},\"user\":{\"last_name\":\"Smith\",\"name\":\"John\"}}",
					),
					resource.TestCheckResourceAttr(
						"commercetools_custom_object.test_nested", "version", "1",
					),
				),
			},
		},
	})
}

func testAccCustomObjectNumber() string {
	return `
	resource "commercetools_custom_object" "test_number" {
		container = "foobar"
		key = "value"
		value = jsonencode(10)
	  }`
}

func testAccCustomObjectNumberUpdateValue() string {
	return `
	resource "commercetools_custom_object" "test_number" {
		container = "foobar"
		key = "value"
		value = jsonencode(20)
	  }`
}

func testAccCustomObjectNumberUpdateKey() string {
	return `
	resource "commercetools_custom_object" "test_number" {
		container = "foobar"
		key = "newvalue"
		value = jsonencode(20)
	  }`
}

func testAccCustomObjectNumberUpdateContainer() string {
	return `
	resource "commercetools_custom_object" "test_number" {
		container = "newbar"
		key = "newvalue"
		value = jsonencode(20)
	  }`
}

func testAccCustomObjectNestedData() string {
	return `
	resource "commercetools_custom_object" "test_nested" {
		container = "foobar"
		key = "nested"
		value = jsonencode({
			address = {
				street = "foo"
				number = 10
			}
			user = {
				name = "John"
				last_name = "Smith"
			}
		})
	  }`
}
