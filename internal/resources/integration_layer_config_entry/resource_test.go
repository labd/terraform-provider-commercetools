package integration_layer_config_entry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

// projectKey is the Project of this test alone.
const projectKey = "acc-cil-config-entry"

func TestAccIntegrationLayerConfigEntry(t *testing.T) {
	name := "commercetools_integration_layer_config_entry.locale"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.TestAccPreCheck(t)
			// Set with the commercetools CLI, outside of Terraform.
			setConfig(t, "EVOLVE_TAX_RATE", "0.21")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			entries := config(t)
			if _, ok := entries["EVOLVE_LOCALE"]; ok {
				return errors.New("EVOLVE_LOCALE is still set after destroy")
			}
			if value(entries["EVOLVE_TAX_RATE"]) != "0.21" {
				return errors.New("an entry Terraform does not manage was removed")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: resources("en-GB", "search-key-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "id", "EVOLVE_LOCALE"),
					resource.TestCheckResourceAttr(name, "value", "en-GB"),
					func(_ *terraform.State) error {
						entries := config(t)
						assert.Equal(t, "en-GB", value(entries["EVOLVE_LOCALE"]))
						assert.False(t, entries["EVOLVE_LOCALE"].Secret)
						// A secret is listed without its value.
						assert.True(t, entries["ALGOLIA_SEARCH_API_KEY"].Secret)
						assert.Empty(t, value(entries["ALGOLIA_SEARCH_API_KEY"]))
						assert.Equal(t, "0.21", value(entries["EVOLVE_TAX_RATE"]))
						return nil
					},
				),
			},
			{
				// A rotated secret is written in place.
				Config: resources("en-GB", "search-key-2"),
				Check:  resource.TestCheckResourceAttr("commercetools_integration_layer_config_entry.algolia_key", "id", "ALGOLIA_SEARCH_API_KEY"),
			},
			{
				// A plain value changed outside of Terraform is set back.
				PreConfig: func() { setConfig(t, "EVOLVE_LOCALE", "de-DE") },
				Config:    resources("en-GB", "search-key-2"),
				Check: func(_ *terraform.State) error {
					assert.Equal(t, "en-GB", value(config(t)["EVOLVE_LOCALE"]))
					return nil
				},
			},
			{
				// Turning a plain value into a secret writes it as one.
				Config: acctest.ProviderConfig(projectKey) + `
resource "commercetools_integration_layer_config_entry" "locale" {
  key          = "EVOLVE_LOCALE"
  secret_value = "en-GB"
}
`,
				Check: func(_ *terraform.State) error {
					entries := config(t)
					assert.True(t, entries["EVOLVE_LOCALE"].Secret)
					_, ok := entries["ALGOLIA_SEARCH_API_KEY"]
					assert.False(t, ok, "the entry left out of the configuration is deleted")
					return nil
				},
			},
			{
				ResourceName:            name,
				ImportState:             true,
				ImportStateId:           "EVOLVE_LOCALE",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_value"},
			},
		},
	})
}

func resources(locale, searchKey string) string {
	return acctest.ProviderConfig(projectKey) + fmt.Sprintf(`
resource "commercetools_integration_layer_config_entry" "locale" {
  key   = "EVOLVE_LOCALE"
  value = %q
}

resource "commercetools_integration_layer_config_entry" "algolia_key" {
  key          = "ALGOLIA_SEARCH_API_KEY"
  secret_value = %q
}
`, locale, searchKey)
}

func client(t *testing.T) *integrationlayer.ByProjectKeyRequestBuilder {
	c, err := acctest.GetIntegrationLayerClient(projectKey)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// config reads the entries by key.
func config(t *testing.T) map[string]integrationlayer.ConfigEntry {
	read, err := client(t).Main().Extension().Config().Get().Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]integrationlayer.ConfigEntry{}
	for _, entry := range read.Entries {
		out[entry.Key] = entry
	}
	return out
}

// setConfig sets a plain entry, as the commercetools CLI would.
func setConfig(t *testing.T, key, value string) {
	_, err := client(t).Main().Extension().Config().Patch(integrationlayer.ExtensionConfigPatch{
		Entries: []integrationlayer.ConfigEntryPatch{{Key: key, Value: value}},
	}).Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func value(entry integrationlayer.ConfigEntry) string {
	text, _ := entry.Value.(string)
	return text
}
