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

func TestAccStore_createAndUpdateWithID(t *testing.T) {

	name := "test method"
	resourceName := "commercetools_store.test"
	key := "test-method"
	languages := []string{"en-US"}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStoreDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStoreDestroy,
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

func TestAccStore_CustomField(t *testing.T) {
	resourceName := "commercetools_store.test"
	name := "test method"
	key := "standard"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStoreDestroy,
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
		return nil, fmt.Errorf("Store not found")
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.Stores().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
