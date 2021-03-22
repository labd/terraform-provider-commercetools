resource "commercetools_project_settings" "project" {
  name = "My project"
  countries = ["NL", "DE", "US", "CA"]
  currencies = ["EUR", "USD", "CAD"]
  languages = ["nl", "de", "en", "fr-CA"]
  external_oauth = {
    url = "https://example.com/oauth/introspection"
    authorization_header = "Bearer secret"
  }
  messages = {
    enabled = true
  }
  carts = {
    country_tax_rate_fallback_enabled = true
  }
}
