resource "commercetools_mcp_server" "storefront_assistant" {
  key                    = "storefront-assistant"
  major_version          = "v1"
  state                  = "Enabled"
  tools                  = ["read_products", "read_categories"]
  tool_output_formatting = "JsonToTabular"
  name = {
    en = "Storefront assistant"
  }

  tool_customization {
    tool        = "read_products"
    description = "Search the product catalog by name or category."
    parameter {
      name         = "limit"
      default_json = jsonencode(20)
    }
  }

  json_output_filtering {
    redacted_with = "***"
    redact {
      field_name_contains = ["secret", "token"]
    }
  }
}

output "mcp_url" {
  value = commercetools_mcp_server.storefront_assistant.url
}
