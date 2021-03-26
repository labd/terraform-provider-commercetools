package commercetools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/labd/commercetools-go-sdk/commercetools"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
	conn := testAccProvider.Meta().(*commercetools.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_tax_category" {
			continue
		}
		response, err := conn.TaxCategoryGetWithID(context.Background(), rs.Primary.ID)
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		// If we don't get a was not found error, return the actual error. Otherwise resource is destroyed
		if !strings.Contains(err.Error(), "was not found") {
			return err
		}
	}
	return nil
}
