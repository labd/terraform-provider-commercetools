# Customer Groups

A Customer can be a member of a customer group (for example reseller, gold member). Special prices can be assigned to specific products based on a customer group.

Also see the [Customer Groups HTTP API documentation](https://docs.commercetools.com/http-api-projects-customerGroups).

## Example Usage

```hcl
resource "commercetools_customer_group" "standard" {
  name = "Standard Customer Group"
  key  = "standard-customer-group"
}

resource "commercetools_customer_group" "golden" {
  name = "Golden Customer Group"
  key  = "golden-customer-group"
}
```

## Argument Reference

* `name` - string - Required
* `key` - string - Optional
