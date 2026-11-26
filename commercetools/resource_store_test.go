package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccStore_createAndUpdateWithID(t *testing.T) {
	name := "test method"
	resourceName := "commercetools_store.test"
	key := "test-method"
	languages := []string{"en-US"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStoreConfig("test", name, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", name),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						res, err := testGetStore(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.EqualValues(t, res.Key, key)
						assert.NotNil(t, res.Name)
						assert.EqualValues(t, (*res.Name)["en"], name)
						return nil
					},
				),
			},
			{
				Config: testAccStoreConfig("test", "new test method", key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "new test method"),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						res, err := testGetStore(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.EqualValues(t, (*res.Name)["en"], "new test method")
						assert.EqualValues(t, res.Languages, []string{})
						return nil
					},
				),
			},
			{
				Config: testAccStoreConfigWithLanguages("test", name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "languages.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "languages.0", "en-US"),
					func(s *terraform.State) error {
						res, err := testGetStore(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.EqualValues(t, res.Languages, []string{"en-US"})
						return nil
					},
				),
			},
			{
				Config: testAccStoreConfigWithCountries("test", name, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "countries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "countries.0", "BE"),
					resource.TestCheckResourceAttr(resourceName, "countries.1", "NL"),
				),
			},
			{
				Config: testAccStoreConfigWithLanguages("other", name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("commercetools_store.other", "languages.#", "1"),
					resource.TestCheckResourceAttr("commercetools_store.other", "languages.0", "en-US"),
					func(s *terraform.State) error {
						res, err := testGetStore(s, "commercetools_store.other")
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.EqualValues(t, res.Languages, []string{"en-US"})
						return nil
					},
				),
			},
		},
	})
}

func TestAccStore_createAndUpdateDistributionLanguages(t *testing.T) {
	resourceName := "commercetools_store.test"
	name := "test dl"
	key := "test-dl"
	languages := []string{"en-US"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStoreConfigWithChannels("test", name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "distribution_channels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution_channels.0", "TEST"),
				),
			},
			{
				Config: testAccStoreConfigWithoutChannels("test", name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "distribution_channels.#", "0"),
				),
			},
			{
				Config: testAccStoreConfigWithChannels("test", name, key, languages),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "distribution_channels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "distribution_channels.0", "TEST"),
				),
			},
		},
	})
}

func TestAccStoreChannelsIgnoreOrder(t *testing.T) {
	resourceName := "commercetools_store.test"
	name := "test channel order"
	key := "test-channel-order"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStoreConfigWithChannelOrder("test", name, key, []string{"first", "second", "third"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "distribution_channels.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution_channels.*", "CHANNEL-FIRST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution_channels.*", "CHANNEL-SECOND"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution_channels.*", "CHANNEL-THIRD"),
					resource.TestCheckResourceAttr(resourceName, "supply_channels.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supply_channels.*", "CHANNEL-FIRST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supply_channels.*", "CHANNEL-SECOND"),
					resource.TestCheckTypeSetElemAttr(resourceName, "supply_channels.*", "CHANNEL-THIRD"),
				),
			},
			{
				// The store keeps the channel order of the first step, so a reordered configuration must plan no changes.
				Config:   testAccStoreConfigWithChannelOrder("test", name, key, []string{"third", "first", "second"}),
				PlanOnly: true,
			},
		},
	})
}

func TestStoreChannelsDiffIgnoresOrder(t *testing.T) {
	testCases := map[string]string{
		"distribution-channels": "distribution_channels",
		"supply-channels":       "supply_channels",
	}

	for testName, attribute := range testCases {
		t.Run(testName, func(t *testing.T) {
			storeResource := resourceStore()

			data := storeResource.TestResourceData()
			data.SetId("store-id")
			require.NoError(t, data.Set("key", "store-key"))
			require.NoError(t, data.Set(attribute, []string{"variant-53", "variant-30", "webshop", "variant-11"}))

			config := terraform.NewResourceConfigRaw(map[string]any{
				"key":     "store-key",
				attribute: []any{"webshop", "variant-11", "variant-30", "variant-53"},
			})

			sameChannels, err := storeResource.Diff(context.Background(), data.State(), config, nil)
			require.NoError(t, err)
			assert.True(t, sameChannels == nil || sameChannels.Empty(), "expected no diff, got %v", sameChannels)

			config = terraform.NewResourceConfigRaw(map[string]any{
				"key":     "store-key",
				attribute: []any{"webshop", "variant-11", "variant-30"},
			})

			removedChannel, err := storeResource.Diff(context.Background(), data.State(), config, nil)
			require.NoError(t, err)
			require.NotNil(t, removedChannel)
			assert.False(t, removedChannel.Empty())
		})
	}
}

// State written while the channel attributes were lists stores them as JSON arrays in API order.
// The set schema must load that state unchanged, because the resource has no state upgrader.
func TestStoreChannelsListStateLoadsAsSet(t *testing.T) {
	listState := []byte(`{
		"id": "store-id",
		"key": "store-key",
		"name": {"en": "Store"},
		"version": 3,
		"languages": ["en-US"],
		"countries": ["NL"],
		"distribution_channels": ["variant-53", "variant-30", "webshop"],
		"supply_channels": ["variant-11", "webshop"],
		"product_selection": [],
		"custom": []
	}`)

	server := schema.NewGRPCProviderServer(New("snapshot")())
	response, err := server.UpgradeResourceState(context.Background(), &tfprotov5.UpgradeResourceStateRequest{
		TypeName: "commercetools_store",
		Version:  0,
		RawState: &tfprotov5.RawState{JSON: listState},
	})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)

	upgraded, err := msgpack.Unmarshal(response.UpgradedState.MsgPack, resourceStore().CoreConfigSchema().ImpliedType())
	require.NoError(t, err)

	expectedChannels := map[string][]string{
		"distribution_channels": {"variant-30", "variant-53", "webshop"},
		"supply_channels":       {"variant-11", "webshop"},
	}
	for attribute, expectedKeys := range expectedChannels {
		channels := upgraded.GetAttr(attribute)
		require.True(t, channels.Type().IsSetType(), "%s is not a set", attribute)

		channelKeys := make([]string, 0, channels.LengthInt())
		for _, channel := range channels.AsValueSlice() {
			channelKeys = append(channelKeys, channel.AsString())
		}
		assert.ElementsMatch(t, expectedKeys, channelKeys)
	}
}

func TestAccStore_CustomField(t *testing.T) {
	resourceName := "commercetools_store.test"
	name := "test method"
	key := "standard"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStoreConfigWithCustomField("test", name, key, []string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", name),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					func(s *terraform.State) error {
						res, err := testGetStore(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.NotNil(t, res.Custom)
						assert.NotNil(t, res.Custom.Fields)
						assert.EqualValues(t, res.Custom.Fields["my-field"], "foobar")
						return nil
					},
				),
			},
			{
				Config: testAccStoreConfigWithChannels("test", name, key, []string{}),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						res, err := testGetStore(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, res)
						assert.Nil(t, res.Custom)
						return nil
					},
				),
			},
		},
	})
}

func testAccStoreConfig(id, name, key string) string {
	return hclTemplate(`
        resource "commercetools_store" "{{ .id }}" {
            key = "{{ .key }}"
            name = {
                en = "{{ .name }}"
                nl = "{{ .name }}"
            }
        }
    `,
		map[string]any{
			"id":   id,
			"name": name,
			"key":  key,
		})
}

func testAccStoreConfigWithLanguages(id, name, key string, languages []string) string {
	return hclTemplate(`
        resource "commercetools_store" "{{ .id }}" {
            key = "{{ .key }}"
            name = {
                en = "{{ .name }}"
                nl = "{{ .name }}"
            }
            languages = {{ .languages | printf "%q" }}
        }
    `, map[string]any{
		"id":        id,
		"name":      name,
		"key":       key,
		"languages": languages,
	})
}

func testAccStoreConfigWithCountries(id, name, key string) string {
	return hclTemplate(`
        resource "commercetools_store" "{{ .id }}" {
            key = "{{ .key }}"
            name = {
                en = "{{ .name }}"
                nl = "{{ .name }}"
            }
            countries = ["BE", "NL"]
        }
    `, map[string]any{
		"id":   id,
		"name": name,
		"key":  key,
	})
}

func testAccStoreConfigWithChannels(id, name, key string, languages []string) string {
	return hclTemplate(`
        resource "commercetools_channel" "{{ .id }}_channel" {
            key = "TEST"
            roles = ["ProductDistribution"]
        }

        resource "commercetools_store" "{{ .id }}" {
            key = "{{ .key }}"
            name = {
                en = "{{ .name }}"
                nl = "{{ .name }}"
            }
            languages = {{ .languages | printf "%q" }}
            distribution_channels = [commercetools_channel.{{ .id }}_channel.key]
        }
    `, map[string]any{
		"id":        id,
		"name":      name,
		"key":       key,
		"languages": languages,
	})
}

// The channels are created in a fixed order, order sets the order in which the store references them.
func testAccStoreConfigWithChannelOrder(id, name, key string, order []string) string {
	return hclTemplate(`
        resource "commercetools_channel" "{{ .id }}_first" {
            key = "CHANNEL-FIRST"
            roles = ["ProductDistribution", "InventorySupply"]
        }

        resource "commercetools_channel" "{{ .id }}_second" {
            key = "CHANNEL-SECOND"
            roles = ["ProductDistribution", "InventorySupply"]
        }

        resource "commercetools_channel" "{{ .id }}_third" {
            key = "CHANNEL-THIRD"
            roles = ["ProductDistribution", "InventorySupply"]
        }

        resource "commercetools_store" "{{ .id }}" {
            key = "{{ .key }}"
            name = {
                en = "{{ .name }}"
            }
            distribution_channels = [
                {{- range .order }}
                commercetools_channel.{{ $.id }}_{{ . }}.key,
                {{- end }}
            ]
            supply_channels = [
                {{- range .order }}
                commercetools_channel.{{ $.id }}_{{ . }}.key,
                {{- end }}
            ]
        }
    `, map[string]any{
		"id":    id,
		"name":  name,
		"key":   key,
		"order": order,
	})
}

func testAccStoreConfigWithoutChannels(id, name, key string, languages []string) string {
	return hclTemplate(`
        resource "commercetools_store" "{{ .id }}" {
            name = {
                en = "{{ .name }}"
                nl = "{{ .name }}"
            }
            key = "{{ .key }}"
            languages = {{ .languages | printf "%q" }}
        }
    `, map[string]any{
		"id":        id,
		"key":       key,
		"name":      name,
		"languages": languages})
}

func testAccStoreConfigWithCustomField(id, name, key string, languages []string) string {
	return hclTemplate(`
        resource "commercetools_type" "{{ .id }}_type" {
            key = "test-for-store"
            name = {
                en = "for Store"
            }
            description = {
                en = "Custom Field for store resource"
            }

            resource_type_ids = ["store"]

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
                name = "localized_string"
                label = {
                    en = "localized string value"
                }
                type {
                    name = "LocalizedString"
                }
            }
            field {
                name = "boolean"
                label = {
                    en = "boolean value"
                }
                type {
                    name = "Boolean"
                }
            }

            field {
                name = "number"
                label = {
                    en = "number value"
                }
                type {
                    name = "Number"
                }
            }

            field {
                name = "money"
                label = {
                    en = "money value"
                }
                type {
                    name = "Money"
                }
            }

            field {
                name = "date"
                label = {
                    en = "date value"
                }
                type {
                    name = "Date"
                }
            }

            field {
                name = "time"
                label = {
                    en = "time value"
                }
                type {
                    name = "Time"
                }
            }

            field {
                name = "datetime"
                label = {
                    en = "datetime value"
                }
                type {
                    name = "DateTime"
                }
            }

            field {
                name = "enum"
                label = {
                    en = "enum value"
                }
                type {
                  name = "Enum"
                  value {
                    key = "day"
                    label = "Daytime"
                  }
                  value {
                    key = "night"
                    label = "Nighttime"
                  }
                }
            }

            field {
                name = "localized_enum"
                label = {
                    en = "localized_enum value"
                }
                type {
                  name = "LocalizedEnum"
                  localized_value {
                    key = "day"
                    label = {
                      en = "Daytime"
                      nl = "Dagtijd"
                    }
                  }
                  localized_value {
                    key = "night"
                    label = {
                      en = "Nighttime"
                      nl = "Nachttijd"
                    }
                  }
                }
            }

            field {
                name = "reference"
                label = {
                    en = "reference value"
                }
                type {
                  name = "Reference"
                  reference_type_id = "store"
                }
            }
        }

        resource "commercetools_channel" "{{ .id }}_channel" {
            key = "TEST"
            roles = ["ProductDistribution"]
        }

        resource "commercetools_store" "{{ .id }}" {
            key = "{{ .key }}"
            name = {
                en = "{{ .name }}"
                nl = "{{ .name }}"
            }
            languages = {{ .languages | printf "%q" }}
            distribution_channels = [commercetools_channel.{{ .id }}_channel.key]
            custom {
                type_id = commercetools_type.{{ .id }}_type.id
                fields = {
                    "my-field" = "foobar"
                    boolean = false
                    number  = 10
                    localized_string = jsonencode({
                        "en-US" : "boo!"
                    })
                    money = jsonencode({
                      "type" : "centPrecision",
                      "currencyCode" : "EUR",
                      "centAmount" : 420,
                      "fractionDigits" : 2
                    })
                    date = "1983-08-23"
                    time = "14:23:32.000"
                    datetime = "2018-10-12T14:00:00.000Z"
                    enum = "day"
                    localized_enum = "night"
                    reference = jsonencode({
                      typeId: "store",
                      id: "doesntexist"
                    })
                }
            }
        }
    `, map[string]any{
		"id":        id,
		"key":       key,
		"name":      name,
		"languages": languages,
	})
}

func testAccCheckStoreDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_store":
			{
				response, err := client.Stores().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("store (%s) still exists", rs.Primary.ID)
					}
					continue
				}

				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
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

func testGetStore(s *terraform.State, identifier string) (*platform.Store, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("store not found")
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.Stores().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
