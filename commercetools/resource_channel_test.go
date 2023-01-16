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

func TestAccChannel_AllFields(t *testing.T) {
	resourceName := "commercetools_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewChannelConfigWithAllFields(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					func(s *terraform.State) error {
						result, err := testGetChannel(s, resourceName)
						if err != nil {
							return err
						}

						expected := &platform.Channel{
							Name: &platform.LocalizedString{
								"en": "Lab Digital",
							},
							Description: &platform.LocalizedString{
								"en": "Lab Digital Office",
							},
							Address: &platform.Address{
								Country:    "NL",
								StreetName: stringRef("Reykjavikstraat"),
								PostalCode: stringRef("3543 KH"),
							},
							GeoLocation: platform.GeoJsonPoint{
								Coordinates: []float64{52.10014028522915, 5.064886641132926},
							},
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Address)
						assert.EqualValues(t, expected.Name, result.Name)
						assert.EqualValues(t, expected.Description, result.Description)
						assert.EqualValues(t, expected.Address, result.Address)
						assert.EqualValues(t, expected.GeoLocation, result.GeoLocation)
						return nil
					},
				),
			},
			{
				Config: testAccNewChannel(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						result, err := testGetChannel(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.EqualValues(t, result.Name, &platform.LocalizedString{})
						assert.EqualValues(t, result.Description, &platform.LocalizedString{})
						assert.Nil(t, result.GeoLocation)
						assert.Nil(t, result.Address)
						assert.Nil(t, result.Custom)
						return nil
					},
				),
			},
			{
				Config: testAccNewChannelConfigWithAllFields(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					func(s *terraform.State) error {
						result, err := testGetChannel(s, resourceName)
						if err != nil {
							return err
						}

						expected := &platform.Channel{
							Name: &platform.LocalizedString{
								"en": "Lab Digital",
							},
							Description: &platform.LocalizedString{
								"en": "Lab Digital Office",
							},
							Address: &platform.Address{
								Country:    "NL",
								StreetName: stringRef("Reykjavikstraat"),
								PostalCode: stringRef("3543 KH"),
							},
							GeoLocation: platform.GeoJsonPoint{
								Coordinates: []float64{52.10014028522915, 5.064886641132926},
							},
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Address)
						assert.EqualValues(t, expected.Name, result.Name)
						assert.EqualValues(t, expected.Description, result.Description)
						assert.EqualValues(t, expected.Address, result.Address)
						assert.EqualValues(t, expected.GeoLocation, result.GeoLocation)
						return nil
					},
				),
			},
		},
	})
}

func TestAccChannel_CustomField(t *testing.T) {
	resourceName := "commercetools_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewChannelConfigWithCustomField(),
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
			{
				Config: testAccNewChannel(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						result, err := testGetChannel(s, resourceName)
						if err != nil {
							return err
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

func testAccNewChannelConfigWithAllFields() string {
	return hclTemplate(`
		resource "commercetools_channel" "test" {
			key = "test"
			roles = ["ProductDistribution"]

			name = {
				en = "Lab Digital"
			}

			description = {
				en = "Lab Digital Office"
			}


			geolocation {
				coordinates = [52.10014028522915, 5.064886641132926]
			}

			address {
				country = "NL"
				street_name = "Reykjavikstraat"
				postal_code = "3543 KH"
			}
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

			field {
				name = "my-enum-set"
				label = {
					en = "My Set of enums"

				}
				type {
					name = "Set"
					element_type {
						name = "Enum"
						value {
							key   = "ENUM-1"
							label = "ENUM 1"
						}
						value {
							key   = "ENUM-2"
							label = "ENUM 2"
						}
						value {
							key   = "ENUM_3"
							label = "ENUM 3"
						}
					}
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
					"my-enum-set" = jsonencode(["ENUM-1", "ENUM-3"])
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

func testGetChannel(s *terraform.State, identifier string) (*platform.Channel, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Channel %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.Channels().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
