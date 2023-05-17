package commercetools

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccAttributeGroups_AllFields(t *testing.T) {
	resourceName := "commercetools_attribute_groups.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAttributeGroupsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewAttributeGroups(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						result, err := testGetAttributeGroups(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.EqualValues(t, "test", result.Name["en"])
						assert.EqualValues(t, "test", (*result.Description)["en"])
						return nil
					},
				),
			},
			{
				Config: testAccAttributeGroupsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name.en", "test updated"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "test updated"),
				),
			},
		},
	})
}

func testAccNewAttributeGroups() string {
	return hclTemplate(`
		resource "commercetools_attribute_groups" "test" {
			key = "test"
			attributes = ["brand"]
            name = {
				en = "test"
			}
            description = {
				en = "test"
			}
		}
	`, map[string]any{})
}

func testAccAttributeGroupsUpdate() string {
	return hclTemplate(`
		resource "commercetools_attribute_groups" "test" {
			key = "test"
			attributes = ["brand", "line"]
            name = {
				en = "test updated"
			}
            description = {
				en = "test updated"
			}
		}
	`, map[string]any{})
}

func testAccCheckAttributeGroupsDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "commercetools_attribute_groups":
			{
				response, err := client.AttributeGroups().WithId(rs.Primary.ID).Get().Execute(context.Background())
				if err == nil {
					if response != nil && response.ID == rs.Primary.ID {
						return fmt.Errorf("attribute group (%s) still exists", rs.Primary.ID)
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

func testGetAttributeGroups(s *terraform.State, identifier string) (*platform.AttributeGroup, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("Attribute group %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.AttributeGroups().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}
