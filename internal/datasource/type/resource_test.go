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
		Steps: []resource.TestStep{{
			Config: testAccConfigCreateCustomField(),
			Check: resource.ComposeTestCheckFunc(
				func(s *terraform.State) error {
					client, err := acctest.GetClient()
					if err != nil {
						return nil
					}
					result, err := client.Types().WithKey("test").Get().Execute(context.Background())
					if err != nil {
						return nil
					}
					assert.NotNil(t, result)
					assert.Equal(t, result.Key, "test")
					assert.Equal(t, result.FieldDefinitions[0].Name, "my-field")
					return nil
				},
			),
		},
			{
				Config: testAccConfigWithCustomFieldBasedOnIDOutOfDataResource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					func(s *terraform.State) error {
						result, err := testGetChannel(s, resourceName)
						if err != nil {
							return err
						}
						client, err := acctest.GetClient()
						if err != nil {
							return nil
						}
						result_type, err := client.Types().WithKey("test").Get().Execute(context.Background())
						if err != nil {
							return nil
						}
						assert.NotNil(t, result)
						assert.Equal(t, result.Custom.Type.ID, result_type.ID)
						assert.Equal(t, result.Custom.Fields["my-field"], "foobar")
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
func testAccConfigCreateCustomField() string {
	return utils.HCLTemplate(`
	resource "commercetools_type" "test"  {
		key = "test"
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
	`, map[string]any{})
}

func testAccConfigWithCustomFieldBasedOnIDOutOfDataResource() string {
	return utils.HCLTemplate(`
		resource "commercetools_type" "test"  {
			key = "test"
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
			key = "test"
		}

		resource "commercetools_channel" "test" {
			key = "test"
			roles = ["ProductDistribution"]
			custom {
				type_id = data.commercetools_type.test_type_with_key.id
				fields = {
					"my-field" = "foobar"
				}
			}
		}

	`, map[string]any{})
}
