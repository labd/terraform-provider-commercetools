package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccCustomerGroupCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCustomerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_customer_group.standard", "name", "Standard name",
					),
					resource.TestCheckResourceAttr(
						"commercetools_customer_group.standard", "key", "standard-key",
					),
				),
			},
			{
				Config: testAccCustomerGroupUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_customer_group.standard", "name", "Standard name new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_customer_group.standard", "key", "standard-key-new",
					),
				),
			},
			{
				Config: testAccCustomerGroupRemoveProperties(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_customer_group.standard", "name", "Standard name new",
					),
					resource.TestCheckResourceAttr(
						"commercetools_customer_group.standard", "key", "",
					),
				),
			},
		},
	})
}

func testAccCustomerGroupConfig() string {
	return `
resource "commercetools_customer_group" "standard" {
	name = "Standard name"
	key  = "standard-key"
}
`
}

func testAccCustomerGroupUpdate() string {
	return `
resource "commercetools_customer_group" "standard" {
	name = "Standard name new"
	key  = "standard-key-new"
}
`
}

func testAccCustomerGroupRemoveProperties() string {
	return `
resource "commercetools_customer_group" "standard" {
	name = "Standard name new"
}
`
}

func testAccCheckCustomerGroupDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_customer_group" {
			continue
		}
		response, err := client.CustomerGroups().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("customer group (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
