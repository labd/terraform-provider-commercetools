# Shipping methods

Shipping methods describe the available shipping methods for orders.

Also see the [shipping methods HTTP API documentation][commercetool-shipping-methods].

## Example Usage

```hcl
resource "commercetools_shipping_method" "standard" {
  name = "Standard tax category"
  key = "Standard tax category"
  description = "Standard tax category"
  is_default = true
  tax_category_id = "<some tax category id>"
}
```

## Argument Reference

The following arguments are supported:

* `name` - Name of the shipping method
* `key` - (Optional) User-specific unique identifier for the shipping method
* `description` - (Optional) Description of the shipping method 
* `is_default` - Whether it should be the default shipping method. There can be only one default shipping method.
* `tax_category_id` - ID to a tax category


### Shipping Rate
A [SubRate][commercetool-subrate] is used to calculate the taxPortions field in a cart or order. It is useful if the total tax of a country is a combination of multiple taxes (e.g. state and local taxes).

These can have the following arguments:

* `name`
* `amount` - Number Percentage in the range of [0..1]


[commercetool-tax-categories]: https://docs.commercetools.com/http-api-projects-taxCategories.html
[commercetool-rate]: https://docs.commercetools.com/http-api-projects-taxCategories.html#taxrate
[commercetool-subrate]: https://docs.commercetools.com/http-api-projects-taxCategories.html#subrate
[country-iso]: https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2
