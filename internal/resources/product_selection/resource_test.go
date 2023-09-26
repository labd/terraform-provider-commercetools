package product_selection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestProductSelctionResource_Create(t *testing.T) {
	rn := "commercetools_product_selection.test_product_selection"

	id := "test_product_selection"
	key := "ps-1"
	name := "the selection"
	mode := "Individual"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testProductSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testProductSelectionConfig(id, name, key, mode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "name", name),
					resource.TestCheckResourceAttr(rn, "key", key),
					resource.TestCheckResourceAttr(rn, "mode", mode),
				),
			},
			{
				Config: testProductSelectionConfigUpdate(id, "the selection updated", key, mode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "name", "the selection updated"),
					resource.TestCheckResourceAttr(rn, "key", key),
				),
			},
		},
	})
}

func testProductSelectionDestroy(_ *terraform.State) error {
	return nil
}

func testProductSelectionConfig(identifier, name, key string, mode string) string {
	return utils.HCLTemplate(`
		resource "commercetools_product_selection" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = "{{ .name }}"
			mode = "{{ .mode }}"
		}
	`, map[string]any{
		"identifier": identifier,
		"name":       name,
		"key":        key,
		"mode":       mode,
	})
}

func testProductSelectionConfigUpdate(identifier, name, key string, mode string) string {
	return utils.HCLTemplate(`
		resource "commercetools_product_selection" "{{ .identifier }}" {
			key = "{{ .key }}"
			name = "{{ .name }}"
			mode = "{{ .mode }}"
		}
	`, map[string]any{
		"identifier": identifier,
		"name":       name,
		"key":        key,
		"mode":       mode,
	})
}
