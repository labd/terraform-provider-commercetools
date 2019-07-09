package commercetools

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTaxCategory_createAndUpdateWithID(t *testing.T) {

	name := "test category"
	key := "test-category"
	description := "test category description"

	newName := "new test category"
	newKey := "new-test-category"
	newDescription := "new test category description"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTaxCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxCategoryConfig(name, key, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "name", name,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "key", key,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "description", description,
					),
				),
			},
			{
				Config: testAccTaxCategoryConfig(newName, newKey, newDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "name", newName,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "key", newKey,
					),
					resource.TestCheckResourceAttr(
						"commercetools_tax_category.standard", "description", newDescription,
					),
				),
			},
		},
	})
}

func testAccTaxCategoryConfig(name string, key string, description string) string {
	return fmt.Sprintf(`
resource "commercetools_tax_category" "standard" {
	name = "%s"
	key = "%s"
	description = "%s"
}`, name, key, description)
}

func testAccCheckTaxCategoryDestroy(s *terraform.State) error {
	return nil
}
