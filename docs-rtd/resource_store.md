# Stores

Stores let you model the context your customers shop in.

Also see the [stores HTTP API documentation][commercetool-stores].

## Example Usage

```hcl
resource "commercetools_store" "standard" {
  name = {
      nl-NL = "My standard store"
  }
  key = "standard-store"

  // optional
  languages            = ["nl-NL"]
  distribution_channels = ["NL-DIST"]
  supply_channels = ["NL-SUP"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - Name of the store.
* `key`  - User-specific unique identifier for the store. The key is mandatory and immutable. It is used to reference the store.
* `languages` - Optional array of languages.
* `distribution_channels` - Optional array of distribution channel keys used for [product projection store filtering](https://docs.commercetools.com/http-api-projects-productProjections#prices-beta).
* `supply_channels` - Optional array of supply channel keys used for [product projection store filtering](https://docs.commercetools.com/http-api-projects-productProjections#prices-beta).


[commercetool-stores]: https://docs.commercetools.com/http-api-projects-stores.html
