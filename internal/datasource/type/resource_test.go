package custom_type_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/labd/commercetools-go-sdk/platform"
	acctest "github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccChannel_CustomFieldWithKey(t *testing.T) {
	resourceName := "commercetools_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigWithCustomFieldBasedOnIDOutOfDataResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					func(s *terraform.State) error {
						result, err := testGetChannel(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Custom)
						assert.NotNil(t, result.Custom.Fields)
						assert.EqualValues(t, result.Custom.Fields["my-field"], "foobar")
						assert.EqualValues(t, result.Custom.Fields["my-enum-set"], []any{"ENUM-1", "ENUM-3"})
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	return nil
}
func testGetChannel(s *terraform.State, identifier string) (*platform.Channel, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Channel %s not found", identifier)
	}

	client, err := acctest.GetClient()
	if err != nil {
		return nil, err
	}
	result, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func testAccConfigWithCustomFieldBasedOnIDOutOfDataResource() string {
	return utils.HCLTemplate(`
		resource "commercetools_type" "test"  {
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

		data "commercetools_type" "test_type_with_key"{
			key = "test-for-channel"
		}
		resource "commercetools_channel" "test" {
			key = "test"
			roles = ["ProductDistribution"]
			custom {
				type_id = data.test_type_with_key.key
				fields = {
					"my-field" = "foobar"
				}
			}
		}

	`, map[string]any{})
}
