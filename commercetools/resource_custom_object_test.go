package commercetools

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// )

// func TestAccCustomObjectCreate_basic(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckCustomObjectDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCustomObjectNumber(),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "container", "foobar",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "key", "value",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "value", "{\"number\":10}",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "version", "1",
// 					),
// 				),
// 			},
// 			{
// 				Config: testAccCustomObjectNumberUpdateValue(),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "container", "foobar",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "key", "value",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "value", "{\"number\":20}",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "version", "2",
// 					),
// 				),
// 			},
// 			{
// 				Config: testAccCustomObjectNumberUpdateKey(),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "container", "foobar",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "key", "newvalue",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "value", "{\"number\":20}",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "version", "1",
// 					),
// 				),
// 			},
// 			{
// 				Config: testAccCustomObjectNumberUpdateContainer(),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "container", "newbar",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "key", "newvalue",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "value", "{\"number\":20}",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_number", "version", "1",
// 					),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccCustomObjectCreate_object(t *testing.T) {

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckCustomObjectDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCustomObjectNestedData(),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_nested", "container", "foobar",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_nested", "key", "nested",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_nested", "value", "{\"address\":{\"number\":10,\"street\":\"foo\"},\"user\":{\"last_name\":\"Smith\",\"name\":\"John\"}}",
// 					),
// 					resource.TestCheckResourceAttr(
// 						"commercetools_custom_object.test_nested", "version", "1",
// 					),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccCustomObjectNumber() string {
// 	return `
// 	resource "commercetools_custom_object" "test_number" {
// 		container = "foobar"
// 		key = "value"
// 		value = jsonencode({
// 			number = 10
// 		})
// 	  }`
// }

// func testAccCustomObjectNumberUpdateValue() string {
// 	return `
// 	resource "commercetools_custom_object" "test_number" {
// 		container = "foobar"
// 		key = "value"
// 		value = jsonencode({
// 			number = 20
// 		})
// 	  }`
// }

// func testAccCustomObjectNumberUpdateKey() string {
// 	return `
// 	resource "commercetools_custom_object" "test_number" {
// 		container = "foobar"
// 		key = "newvalue"
// 		value = jsonencode({
// 			number = 20
// 		})
// 	  }`
// }

// func testAccCustomObjectNumberUpdateContainer() string {
// 	return `
// 	resource "commercetools_custom_object" "test_number" {
// 		container = "newbar"
// 		key = "newvalue"
// 		value = jsonencode({
// 			number = 20
// 		})
// 	  }`
// }

// func testAccCustomObjectNestedData() string {
// 	return `
// 	resource "commercetools_custom_object" "test_nested" {
// 		container = "foobar"
// 		key = "nested"
// 		value = jsonencode({
// 			address = {
// 				street = "foo"
// 				number = 10
// 			}
// 			user = {
// 				name = "John"
// 				last_name = "Smith"
// 			}
// 		})
// 	  }`
// }

// func testAccCheckCustomObjectDestroy(s *terraform.State) error {
// 	conn := getClient(testAccProvider.Meta())

// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "commercetools_custom_object" {
// 			continue
// 		}
// 		container := rs.Primary.Attributes["container"]
// 		response, err := conn.CustomObjects().WithContainer(container).Get().Execute(context.Background())
// 		if err == nil {
// 			if response != nil && response.Count > 0 {
// 				return fmt.Errorf("custom object container (%s) still exists", container)
// 			}
// 			return nil
// 		}

// 		if newErr := checkApiResult(err); newErr != nil {
// 			return newErr
// 		}
// 	}
// 	return nil
// }
