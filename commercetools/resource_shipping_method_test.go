package commercetools

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccShippingMethod_createAndUpdateWithID(t *testing.T) {

	name := "test sh method"
	key := "test-sh-method"
	description := "test shipping method description"
	localizedName := "some localized shipping method test name"
	predicate := "1 = 1"
	resourceName := "commercetools_shipping_method.standard"

	newName := "new test sh method"
	newKey := "new-test-sh-method"
	newDescription := "new test shipping method description"
	newLocalizedName := "some new localized shipping method test name"
	newPredicate := "2 = 2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodConfig(name, key, description, description, localizedName, false, true, predicate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "localized_description.en", description),
					resource.TestCheckResourceAttr(resourceName, "localized_name.en", localizedName),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(resourceName, "predicate", predicate),
				),
			},
			{
				Config: testAccShippingMethodConfig(newName, newKey, newDescription, newDescription, newLocalizedName, true, true, newPredicate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", newName),
					resource.TestCheckResourceAttr(resourceName, "key", newKey),
					resource.TestCheckResourceAttr(resourceName, "description", newDescription),
					resource.TestCheckResourceAttr(resourceName, "localized_description.en", newDescription),
					resource.TestCheckResourceAttr(resourceName, "localized_name.en", newLocalizedName),
					resource.TestCheckResourceAttr(resourceName, "is_default", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "tax_category_id"),
					resource.TestCheckResourceAttr(resourceName, "predicate", newPredicate),
				),
			},
		},
	})
}

func testAccShippingMethodConfig(name string, key string, description string, localizedDescription string, localizedName string, isDefault bool, setTaxCategory bool, predicate string) string {
	taxCategoryReference := ""
	if setTaxCategory {
		taxCategoryReference = "tax_category_id = commercetools_tax_category.test.id"
	}
	return hclTemplate(`
		resource "commercetools_tax_category" "test" {
			name = "test"
			key = "test"
			description = "test"
		}

		resource "commercetools_shipping_method" "standard" {
			name = "{{ .name }}"
			key = "{{ .key }}"
			description = "{{ .description }}"
			localized_description = {
				en = "{{ .localizedDescription }}"
			}
			localized_name = {
				en = "{{ .localizedName }}"
			}
			is_default = "{{ .isDefault }}"
			predicate = "{{ .predicate }}"

			{{ .taxCategoryReference }}
		}
		`,
		map[string]any{
			"name":                 name,
			"key":                  key,
			"description":          description,
			"localizedDescription": localizedDescription,
			"localizedName":        localizedName,
			"isDefault":            isDefault,
			"predicate":            predicate,
			"taxCategoryReference": taxCategoryReference,
		})
}

func testAccCheckShippingMethodDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_tax_category":
			{
				response, err := client.TaxCategories().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("tax category (%s) still exists", rs.Primary.ID)
					}
					continue
				}

				if newErr := checkApiResult(err); newErr != nil {
					return newErr
				}
			}
		case "commercetools_shipping_method":
			{
				response, err := client.ShippingMethods().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("shipping method (%s) still exists", rs.Primary.ID)
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
