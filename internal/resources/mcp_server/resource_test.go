package mcp_server_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

func TestAccMcpServerCreate_basic(t *testing.T) {
	resourceName := "commercetools_mcp_server.assistant"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMcpServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
resource "commercetools_mcp_server" "assistant" {
  key           = "storefront-assistant"
  major_version = "v1"
  tools         = ["read_products", "read_categories"]
  name = {
    en = "Storefront assistant"
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "storefront-assistant"),
					resource.TestCheckResourceAttr(resourceName, "state", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode", "ClientCredentials"),
					resource.TestCheckResourceAttr(resourceName, "type", "CommerceMCP"),
					resource.TestCheckResourceAttr(resourceName, "tools.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
			{
				Config: `
resource "commercetools_mcp_server" "assistant" {
  key                    = "storefront-assistant"
  major_version          = "v1"
  state                  = "Enabled"
  tools                  = ["read_products", "read_orders"]
  tool_output_formatting = "JsonToTabular"
  name = {
    en = "Storefront assistant"
  }

  tool_customization {
    tool        = "read_products"
    description = "Products in the catalog"
    parameter {
      name         = "limit"
      default_json = jsonencode(20)
    }
  }

  json_output_filtering {
    redacted_with = "***"
    drop {
      field_names = ["password"]
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "Enabled"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tools.*", "read_orders"),
					resource.TestCheckResourceAttr(resourceName, "tool_customization.0.parameter.0.default_json", "20"),
					resource.TestCheckResourceAttr(resourceName, "json_output_filtering.0.drop.0.field_names.0", "password"),
					resource.TestCheckResourceAttr(resourceName, "tool_output_formatting", "JsonToTabular"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMcpServerDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_mcp_server" {
			continue
		}
		response, err := client.McpServers().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("MCP server (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
