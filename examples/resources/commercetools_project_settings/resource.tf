resource "commercetools_project_settings" "my-project" {
  name       = "My project"
  countries  = ["NL", "DE", "US", "CA"]
  currencies = ["EUR", "USD", "CAD"]
  languages  = ["nl", "de", "en", "fr-CA"]
  external_oauth {
    url                  = "https://example.com/oauth/introspection"
    authorization_header = "Bearer secret"
  }
  messages {
    enabled = true
  }
  carts {
    country_tax_rate_fallback_enabled   = false
    delete_days_after_last_modification = 10
    price_rounding_mode                 = "HalfUp"
    tax_rounding_mode                   = "HalfUp"
  }

  shopping_lists {
    delete_days_after_last_modification = 100
  }

  shipping_rate_input_type = "CartClassification"

  enable_search_index_products       = true
  enable_search_index_orders         = true
  enable_search_index_customers      = true
  enable_search_index_business_units = true

  shipping_rate_cart_classification_value {
    key = "Small"
    label = {
      "en" = "Small"
      "nl" = "Klein"
    }
  }
}
