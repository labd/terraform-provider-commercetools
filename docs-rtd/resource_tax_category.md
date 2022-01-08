# Tax Categories

Tax Categories define how products are to be taxed in different countries.

Also see the [tax categories HTTP API documentation][commercetool-tax-categories].

## Example Usage

```hcl
resource "commercetools_tax_category" "standard" {
  name = "Standard tax category"
}
```

## Argument Reference

The following arguments are supported:

* `name` - Name of the tax category
* `key` - (Optional) User-specific unique identifier for the category
* `description` - (Optional) Description of the tax category

[commercetool-tax-categories]: https://docs.commercetools.com/http-api-projects-taxCategories.html