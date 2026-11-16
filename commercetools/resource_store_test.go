package commercetools

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
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

func testProductSelectionSet(items ...map[string]any) *schema.Set {
	elem := resourceStore().Schema["product_selection"].Elem.(*schema.Resource)
	set := schema.NewSet(schema.HashResource(elem), nil)
	for _, item := range items {
		set.Add(item)
	}
	return set
}

func TestExpandProductSelections(t *testing.T) {
	t.Run("by id", func(t *testing.T) {
		got, err := expandProductSelections(testProductSelectionSet(
			map[string]any{"active": true, "product_selection_id": "abc", "product_selection_key": ""},
		))
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "abc", *got[0].ProductSelection.ID)
		assert.Nil(t, got[0].ProductSelection.Key)
		assert.True(t, *got[0].Active)
	})

	t.Run("by key", func(t *testing.T) {
		got, err := expandProductSelections(testProductSelectionSet(
			map[string]any{"active": false, "product_selection_id": "", "product_selection_key": "my-key"},
		))
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Nil(t, got[0].ProductSelection.ID)
		assert.Equal(t, "my-key", *got[0].ProductSelection.Key)
		assert.False(t, *got[0].Active)
	})

	t.Run("neither", func(t *testing.T) {
		_, err := expandProductSelections(testProductSelectionSet(
			map[string]any{"active": true, "product_selection_id": "", "product_selection_key": ""},
		))
		assert.Error(t, err)
	})

	t.Run("both", func(t *testing.T) {
		_, err := expandProductSelections(testProductSelectionSet(
			map[string]any{"active": true, "product_selection_id": "abc", "product_selection_key": "my-key"},
		))
		assert.Error(t, err)
	})
}

func TestFlattenProductSelections(t *testing.T) {
	selections := []platform.ProductSelectionSetting{
		{
			Active: true,
			ProductSelection: platform.ProductSelectionReference{
				ID:  "id-1",
				Obj: &platform.ProductSelection{ID: "id-1", Key: utils.StringRef("key-1")},
			},
		},
		{
			Active: false,
			ProductSelection: platform.ProductSelectionReference{
				ID:  "id-2",
				Obj: &platform.ProductSelection{ID: "id-2", Key: utils.StringRef("key-2")},
			},
		},
		{
			// Not expanded and without a key: resolved via the resolver, falls back to the id.
			Active:           true,
			ProductSelection: platform.ProductSelectionReference{ID: "id-3"},
		},
	}

	// Prior state referenced key-1 by key, key-2 and id-3 by id.
	prior := testProductSelectionSet(
		map[string]any{"active": true, "product_selection_id": "", "product_selection_key": "key-1"},
		map[string]any{"active": false, "product_selection_id": "id-2", "product_selection_key": ""},
	)

	got, err := flattenProductSelections(context.Background(), prior, selections, func(_ context.Context, id string) (string, error) {
		// Only the unexpanded id-3 reaches the resolver; it has no key.
		assert.Equal(t, "id-3", id)
		return "", nil
	})
	assert.NoError(t, err)
	assert.Len(t, got, 3)

	assert.Equal(t, "key-1", got[0]["product_selection_key"])
	assert.NotContains(t, got[0], "product_selection_id")

	assert.Equal(t, "id-2", got[1]["product_selection_id"])
	assert.NotContains(t, got[1], "product_selection_key")

	assert.Equal(t, "id-3", got[2]["product_selection_id"])
	assert.NotContains(t, got[2], "product_selection_key")
}

func TestFlattenProductSelections_UnexpandedIDConfiguredByID(t *testing.T) {
	// One selection is configured by key and one by id, and the API returns
	// neither expanded. Only the key-based one may trigger a lookup; a failing
	// lookup must not be reached for the id-based selection.
	selections := []platform.ProductSelectionSetting{
		{Active: true, ProductSelection: platform.ProductSelectionReference{ID: "id-key"}},
		{Active: true, ProductSelection: platform.ProductSelectionReference{ID: "id-plain"}},
	}
	prior := testProductSelectionSet(
		map[string]any{"active": true, "product_selection_id": "", "product_selection_key": "my-key"},
		map[string]any{"active": true, "product_selection_id": "id-plain", "product_selection_key": ""},
	)

	var looked []string
	got, err := flattenProductSelections(context.Background(), prior, selections, func(_ context.Context, id string) (string, error) {
		looked = append(looked, id)
		if id == "id-plain" {
			return "", errors.New("lookup must not happen for id-configured selections")
		}
		return "my-key", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"id-key"}, looked)
	assert.Equal(t, "my-key", got[0]["product_selection_key"])
	assert.Equal(t, "id-plain", got[1]["product_selection_id"])
	assert.NotContains(t, got[1], "product_selection_key")
}

func TestFlattenProductSelections_UnexpandedKey(t *testing.T) {
	selections := []platform.ProductSelectionSetting{
		{Active: true, ProductSelection: platform.ProductSelectionReference{ID: "id-1"}},
		{Active: true, ProductSelection: platform.ProductSelectionReference{ID: "id-2"}},
	}
	prior := testProductSelectionSet(
		map[string]any{"active": true, "product_selection_id": "", "product_selection_key": "key-1"},
		map[string]any{"active": true, "product_selection_id": "id-2", "product_selection_key": ""},
	)

	t.Run("resolves key", func(t *testing.T) {
		resolve := func(_ context.Context, id string) (string, error) {
			return map[string]string{"id-1": "key-1", "id-2": "key-2"}[id], nil
		}
		got, err := flattenProductSelections(context.Background(), prior, selections, resolve)
		assert.NoError(t, err)
		assert.Equal(t, "key-1", got[0]["product_selection_key"])
		assert.NotContains(t, got[0], "product_selection_id")
		assert.Equal(t, "id-2", got[1]["product_selection_id"])
		assert.NotContains(t, got[1], "product_selection_key")
	})

	t.Run("resolver error is returned", func(t *testing.T) {
		resolve := func(_ context.Context, id string) (string, error) {
			return "", fmt.Errorf("boom")
		}
		_, err := flattenProductSelections(context.Background(), prior, selections, resolve)
		assert.Error(t, err)
	})
}

func TestValidateProductSelectionsConfig(t *testing.T) {
	str := cty.StringVal
	null := cty.NullVal(cty.String)
	unknown := cty.UnknownVal(cty.String)
	block := func(id, key cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"active":                cty.True,
			"product_selection_id":  id,
			"product_selection_key": key,
		})
	}
	config := func(blocks ...cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{"product_selection": cty.SetVal(blocks)})
	}

	tests := []struct {
		name    string
		config  cty.Value
		wantErr bool
	}{
		{"id only", config(block(str("a"), null)), false},
		{"key only", config(block(null, str("k"))), false},
		{"empty string counts as unset", config(block(str(""), str("k"))), false},
		{"neither", config(block(null, null)), true},
		{"neither, empty strings", config(block(str(""), str(""))), true},
		{"both", config(block(str("a"), str("k"))), true},
		{"unknown id, key null", config(block(unknown, null)), false},
		{"unknown key, id null", config(block(null, unknown)), false},
		{"unknown id, key set", config(block(unknown, str("k"))), false},
		{"both unknown", config(block(unknown, unknown)), false},
		{"valid and invalid blocks", config(block(str("a"), null), block(null, null)), true},
		{"unknown set", cty.ObjectVal(map[string]cty.Value{
			"product_selection": cty.UnknownVal(cty.Set(cty.Object(map[string]cty.Type{
				"active":                cty.Bool,
				"product_selection_id":  cty.String,
				"product_selection_key": cty.String,
			}))),
		}), false},
		{"null config", cty.NullVal(cty.EmptyObject), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProductSelectionsConfig(tt.config)
			if tt.wantErr {
				assert.ErrorContains(t, err, "exactly one of product_selection_id or product_selection_key")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAccStore_ProductSelectionRequiresIDOrKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccStoreConfigWithProductSelection("test-ps-invalid", ``, true),
				ExpectError: regexp.MustCompile(`exactly one of[\s│]+product_selection_id[\s│]+or[\s│]+product_selection_key`),
			},
		},
	})
}

func testAccStoreConfigWithProductSelection(key, reference string, active bool) string {
	return hclTemplate(`
        resource "commercetools_store" "test" {
            key = "{{ .key }}"
            name = {
                en = "test store"
            }
            product_selection {
                {{ .reference }}
                active = {{ .active }}
            }
        }
    `,
		map[string]any{
			"key":       key,
			"reference": reference,
			"active":    active,
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
