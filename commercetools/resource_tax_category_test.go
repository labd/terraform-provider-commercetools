package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTaxCategory_createAndUpdateWithID(t *testing.T) {
	resourceName := "commercetools_tax_category.standard"
	name := "test category"
	key := "test-category"
	description := "test category description"

	newName := "new test category"
	newKey := "new-test-category"
	newDescription := "new test category description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckTaxCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxCategoryConfig(name, key, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccTaxCategoryConfig(newName, newKey, newDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", newName),
					resource.TestCheckResourceAttr(resourceName, "key", newKey),
					resource.TestCheckResourceAttr(resourceName, "description", newDescription),
				),
			},
		},
	})
}

func testAccTaxCategoryConfig(name, key, description string) string {
	return hclTemplate(`
		resource "commercetools_tax_category" "standard" {
			name        = "{{ .name }}"
			key         = "{{ .key }}"
			description = "{{ .description }}"
		}
	`, map[string]any{
		"key":         key,
		"name":        name,
		"description": description,
	})

}

func testAccCheckTaxCategoryDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_tax_category" {
			continue
		}
		response, err := client.TaxCategories().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
