# API Extension

Provides a commercetools extension

## Example Usage

```hcl
resource "commercetools_api_extension" "my-extension" {
  key = "test-case"

  destination {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions          = ["Create", "Update"]
  }
}

```

## Argument Reference

The following arguments are supported:

* `key` - User-specific unique identifier for the subscription
* `destination` - Details where the extension can be reached
* `triggers` - Describes what triggers the extension
