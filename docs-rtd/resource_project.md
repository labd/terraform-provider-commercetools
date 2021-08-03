# Project Settings

Lets you change the settings of a commercetools project.

!!! note
    The project itself needs to be set up already. Before you can apply
    changes, you need to import the project:

    ```$ terraform import commercetools_project.project my-project-key```

Also, the project can not be destroyed with terraform.

## Example Usage

```hcl
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
  shipping_rate_input_type = "CartClassification"
  
  shipping_rate_cart_classification_values {
    key = "Small"
    label = {
      "en" = "Small"
      "nl" = "Klein"
    }
  }
  
  
}
```

## Argument Reference

The following arguments are supported:

* `name` -  The name of the project
* `countries` - A two-digit country code as per ISO 3166-1 alpha-2
* `currencies` - A three-digit currency code as per ISO 4217
* `languages` - An IETF language tag
* `external_oauth.url` - The URL for your token introspection endpoint
* `external_oauth.authorization_header` - The authorization header to send when querying the `external_oauth.url`
* `messages.enabled` - When `true` the creation of messages is enabled
* `carts.country_tax_rate_fallback_enabled` - When `true` uses country - _no state_ tax rate fallback when a shipping address state is not explicitly covered in the rates lists of all tax categories of a cart's line items.
* `shipping_rate_input_type` - Allows for three ways to dynamically select a ShippingRatePriceTier: `"CartValue"`, `"CartScore"` and `"CartClassification"`
* `shipping_rate_cart_classification_value` - If shipping_rate_input_type is set to CartClassification these values are used to create tiers