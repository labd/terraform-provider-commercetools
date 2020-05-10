package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccCustomerGroupCreate_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
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
