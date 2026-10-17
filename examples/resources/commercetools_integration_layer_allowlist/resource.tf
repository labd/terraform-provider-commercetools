resource "commercetools_integration_layer_allowlist" "allowlist" {
  allow = [
    "cdn.contentful.com",
    "preview.contentful.com",
    "*.algolia.net",
    "*.algolianet.com",
  ]
}
