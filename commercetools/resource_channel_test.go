package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccChannel_CustomField(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewChannelConfigWithCustomField(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_channel.test", "key", "test",
					),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["commercetools_channel.test"]
						if !ok {
							return fmt.Errorf("Channel not found")
						}

						client := getClient(testAccProvider.Meta())
						result, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
						if err != nil {
							return err
						}
						if result == nil {
							return fmt.Errorf("resource not found")
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Custom)
						assert.NotNil(t, result.Custom.Fields)
						assert.EqualValues(t, result.Custom.Fields["my-field"], "foobar")
						return nil
					},
				),
			},
			{
				Config: testAccNewChannel(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["commercetools_channel.test"]
						if !ok {
							return fmt.Errorf("Channel not found")
						}

						client := getClient(testAccProvider.Meta())
						result, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
						if err != nil {
							return err
						}
						if result == nil {
							return fmt.Errorf("resource not found")
						}

						assert.NotNil(t, result)
						assert.Nil(t, result.Custom)
						return nil
					},
				),
			},
		},
	})
}

func testAccNewChannel() string {
	return hclTemplate(`
		resource "commercetools_channel" "test" {
			key = "test"
			roles = ["ProductDistribution"]
		}
	`, map[string]any{})
}

func testAccNewChannelConfigWithCustomField() string {
	return hclTemplate(`
		resource "commercetools_type" "test" {
			key = "test-for-channel"
			name = {
				en = "for channel"
			}
			description = {
				en = "Custom Field for channel resource"
			}

			resource_type_ids = ["channel"]

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

		resource "commercetools_channel" "test" {
			key = "test"
			roles = ["ProductDistribution"]
			custom {
				type_id = commercetools_type.test.id
				fields = {
					"my-field" = "foobar"
				}
			}
		}
	`, map[string]any{})
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_channel":
			{
				response, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("supply channel (%s) still exists", rs.Primary.ID)
					}
					continue
				}
				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		default:
			continue
		}
	}
	return nil
}
