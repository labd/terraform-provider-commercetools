package attribute_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccAttributeGroupCreate_basic(t *testing.T) {
	resourceName := "commercetools_attribute_group.acctest_attribute_group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupConfig("acctest_attribute_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "attribute-group-key"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "Attribute group name"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "Attribute group description"),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attribute.0.key", "attribute-key-1"),
				),
			},
			{
				Config: testAccAttributeGroupConfigUpdate("acctest_attribute_group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "new-attribute-group-key"),
					resource.TestCheckResourceAttr(resourceName, "name.en", "New attribute group name"),
					resource.TestCheckResourceAttr(resourceName, "description.en", "New attribute group description"),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "attribute.0.key", "attribute-key-1"),
					resource.TestCheckResourceAttr(resourceName, "attribute.1.key", "attribute-key-2"),
				),
			},
		},
	})
}

func testAccCheckAttributeGroupDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_attribute_group" {
			continue
		}
		response, err := client.AttributeGroups().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("attribute group (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testAccAttributeGroupConfig(identifier string) string {
	return utils.HCLTemplate(`
		resource "commercetools_attribute_group" "{{ .identifier }}" {
			key		   	= "attribute-group-key"
			name       	= {
				"en" 	= "Attribute group name"
			}
			description       	= {
				"en" 	= "Attribute group description"
			}
			
			attribute {
				key 	= "attribute-key-1"
			}
		}`, map[string]any{
		"identifier": identifier,
	})
}

func testAccAttributeGroupConfigUpdate(identifier string) string {
	return utils.HCLTemplate(`
		resource "commercetools_attribute_group" "{{ .identifier }}" {
			key		   	= "new-attribute-group-key"
			name       	= {
				"en" 	= "New attribute group name"
			}
			description       	= {
				"en" 	= "New attribute group description"
			}
			
			attribute {
				key 	= "attribute-key-1"
			}
			
			attribute {
				key 	= "attribute-key-2"
			}
		}`, map[string]any{
		"identifier": identifier,
	})
}
