variable "contentful_access_token" {
  type      = string
  sensitive = true
}

# Read by the extension as ctx.config.CONTENTFUL_SPACE_ID.
resource "commercetools_integration_layer_config_entry" "contentful_space" {
  key   = "CONTENTFUL_SPACE_ID"
  value = "my-space-id"
}

resource "commercetools_integration_layer_config_entry" "contentful_token" {
  key          = "CONTENTFUL_ACCESS_TOKEN"
  secret_value = var.contentful_access_token
}
