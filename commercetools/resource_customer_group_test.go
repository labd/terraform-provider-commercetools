package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
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
					func(s *terraform.State) error {
						res, err := testGetCustomerGroup(s, "commercetools_customer_group.standard")
						if err != nil {
							return err
						}
						assert.NotNil(t, res)
						assert.EqualValues(t, res.Key, stringRef("standard-key"))
						assert.EqualValues(t, res.Name, "Standard name")
						return nil
					},
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
					func(s *terraform.State) error {
						res, err := testGetCustomerGroup(s, "commercetools_customer_group.standard")
						if err != nil {
							return err
						}
						assert.NotNil(t, res)
						assert.EqualValues(t, res.Key, stringRef("standard-key-new"))
						assert.EqualValues(t, res.Name, "Standard name new")
						return nil
					},
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
					func(s *terraform.State) error {
						res, err := testGetCustomerGroup(s, "commercetools_customer_group.standard")
						if err != nil {
							return err
						}
						assert.NotNil(t, res)
						assert.Nil(t, res.Key)
						assert.EqualValues(t, res.Name, "Standard name new")
						return nil
					},
				),
			},
		},
	})
}

func TestAccCustomerGroupCreate_CustomField(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCustomerGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGroupCustomField(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						res, err := testGetCustomerGroup(s, "commercetools_customer_group.standard")
						if err != nil {
							return err
						}
						assert.NotNil(t, res)
						assert.NotNil(t, res.Custom)
						return nil
					},
				),
			},
		},
	})
}

func testAccCustomerGroupConfig() string {
	return hclTemplate(`
		resource "commercetools_customer_group" "standard" {
			name = "Standard name"
			key  = "standard-key"
		}`,
		map[string]any{})
}

func testAccCustomerGroupUpdate() string {
	return hclTemplate(`
		resource "commercetools_customer_group" "standard" {
			name = "Standard name new"
			key  = "standard-key-new"
		}`,
		map[string]any{})
}

func testAccCustomerGroupRemoveProperties() string {
	return hclTemplate(`
		resource "commercetools_customer_group" "standard" {
			name = "Standard name new"
		}`,
		map[string]any{})
}

func testAccCustomerGroupCustomField() string {
	return hclTemplate(`
		resource "commercetools_type" "test" {
			key = "test-for-customer-group"
			name = {
				en = "for customer-group"
			}
			description = {
				en = "Custom Field for customer-group resource"
			}

			resource_type_ids = ["customer-group"]

			field {
				name = "my-field"
				label = {
					en = "My Custom field"
				}
				type {
					name = "String"
				}
			}
		}

		resource "commercetools_customer_group" "standard" {
			name = "Standard name"
			key  = "standard-key"
			custom {
				type_id = commercetools_type.test.id
				fields = {
					"my-field" = "bar"
				}
			}
		}`,
		map[string]any{})
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

func testGetCustomerGroup(s *terraform.State, identifier string) (*platform.CustomerGroup, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("CustomerGroup not found")
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.CustomerGroups().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
