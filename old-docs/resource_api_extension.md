# API Extension

Provides a commercetools API extension

Also see the [extension HTTP API documentation](https://docs.commercetools.com/http-api-projects-api-extensions).

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
* `timeout_in_ms` - The maximum time the commercetools platform waits for a
  response from the extension. If not present, 2000 (2 seconds) is used.

