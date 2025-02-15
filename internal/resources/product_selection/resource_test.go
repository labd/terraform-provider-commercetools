package product_selection_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestProductSelectionResource_Create(t *testing.T) {
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
					resource.TestCheckResourceAttr(rn, "name.en", name),
					resource.TestCheckResourceAttr(rn, "key", key),
					resource.TestCheckResourceAttr(rn, "mode", mode),
				),
			},
			{
				Config: testProductSelectionConfigUpdate(id, "the selection updated", key, mode, "my-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "name.en", "the selection updated"),
					resource.TestCheckResourceAttr(rn, "key", key),
					resource.TestCheckResourceAttrWith(rn, "custom.type_id", acctest.IsValidUUID),
					resource.TestCheckResourceAttr(rn, "custom.fields.my-field", "my-value"),
				),
			},
		},
	})
}

func testProductSelectionDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product_selection" {
			continue
		}
		response, err := client.ProductSelections().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("product selection (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testProductSelectionConfig(identifier, name, key string, mode string) string {
	return utils.HCLTemplate(`
		resource "commercetools_product_selection" "{{ .identifier }}" {
			key = "{{ .key }}"
			name       	= {
				"en" 	= "{{ .name }}"
			}
			mode = "{{ .mode }}"
		}
	`, map[string]any{
		"identifier": identifier,
		"name":       name,
		"key":        key,
		"mode":       mode,
	})
}

func testProductSelectionConfigUpdate(identifier, name, key, mode, customValue string) string {
	return utils.HCLTemplate(`
		resource "commercetools_type" "my-type-{{ .identifier }}" {
		  key = "my-type"
		  name = {
			en = "My type"
			nl = "Mijn type"
		  }
		
		  resource_type_ids = ["product-selection"]
		
		  field {
			name = "my-field"
			label = {
			  en = "My field"
			  nl = "Mijn veld"
			}
			type {
			  name = "String"
			}
		  }
		}
	
		resource "commercetools_product_selection" "{{ .identifier }}" {
			key = "{{ .key }}"
			name       	= {
				"en" 	= "{{ .name }}"
			}
			mode = "{{ .mode }}"
	
		   custom {
			 type_id = commercetools_type.my-type-{{ .identifier }}.id
			 fields = {
			   my-field = "{{ .custom_value }}"
			 } 
		   }
		}
	`, map[string]any{
		"identifier":   identifier,
		"name":         name,
		"key":          key,
		"mode":         mode,
		"custom_value": customValue,
	})
}
