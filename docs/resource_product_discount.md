# Product Discounts

A product discount applies to a specific product or subset of products based criteria you provide.

## Example Usage

```hcl
resource "commercetools_product_discount" "10_percent_off" {
  name = "10% off"
  key  = "10-percent-off"
  description = 
  rate {
    name = "21% BTW"
    amount = 0.21
    country = "NL"
    included_in_price = false
  }
  rate {
    name = "5% US"
    amount = 0.05
    country = "US"
    included_in_price = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - Name of the tax category
* `key` - (Optional) User-specific unique identifier for the category
* `description` - (Optional) Description of the tax category
* `rate` - Can be 1 or more [rates](#rates)


### Rates
[Tax Rates][commercetool-rate] defines specific rates.

These can have the following arguments:

* `name` - Tax rate name
* `amount` - Number Percentage in the range of [0..1]. The sum of the amounts of all sub rates, if there are any. Optional if sub_rates are defined, or equal to the sum of all sub_rates.
* `include_in_price`
* `country` - A two-digit country code as per [ISO 3166-1 alpha-2][country-iso]
* `state` - (Optional) The state in the country
* `sub_rate` - Can be 1 or more [subrates](#sub-rates)


### Sub rates
A [SubRate][commercetool-subrate] is used to calculate the taxPortions field in a cart or order. It is useful if the total tax of a country is a combination of multiple taxes (e.g. state and local taxes).

These can have the following arguments:

* `name`
* `amount` - Number Percentage in the range of [0..1]


[commercetool-tax-categories]: https://docs.commercetools.com/http-api-projects-taxCategories.html
[commercetool-rate]: https://docs.commercetools.com/http-api-projects-taxCategories.html#taxrate
[commercetool-subrate]: https://docs.commercetools.com/http-api-projects-taxCategories.html#subrate
[country-iso]: https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2
