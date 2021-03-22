# Tax Category Rates

Tax Category Rates define specify tax rates for products in different countries/states.

Also see the [tax categories HTTP API documentation][commercetool-tax-categories].

## Example Usage

```hcl
resource "commercetools_tax_category" "standard" {
  name = "Standard tax category"
  key  = "standard-tax-category"
}

resource "commercetools_tax_category_rate" "standard-tax-category-DE" {
  tax_category_id   = "${commercetools_tax_category.standard.id}"
  name              = "19% MwSt"
  amount            = 0.19
  included_in_price = false
  country           = "DE"
}

resource "commercetools_tax_category_rate" "standard-tax-category-NL" {
  tax_category_id   = "${commercetools_tax_category.standard.id}"
  name              = "21% BTW"
  amount            = 0.21
  included_in_price = true
  country           = "NL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - Tax rate name
* `amount` - Number Percentage in the range of [0..1]. The sum of the amounts of all sub rates, if there are any. If sub_rates are defined, it should be equal to the sum of all sub_rates.
* `include_in_price` - Boolean
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
