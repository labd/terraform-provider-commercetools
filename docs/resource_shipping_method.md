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
  predicate = "1 = 1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - Name of the shipping method.
* `key` - (Optional) User-specific unique identifier for the shipping method.
* `description` - (Optional) Description of the shipping method.
* `is_default` - Whether it should be the default shipping method. There can be only one default shipping method.
* `tax_category_id` - ID to a tax category.
* `predicate` - Predicate conditions for shipping method aligibility. 


### Shipping Zone Rate *BETA, subject to changes*
A [Shipping Zone Rate][commercetool-shipping-zone-rate] is used to set shipping costs per zone per currency.

These can have the following arguments:

* `shipping_method_id` - Id of the shipping method.
* `shipping_zone_id` - Id of the shipping zone.
* `price` - Single entry configuring the price of the shipping cost to the specified zone.
* `free_above` - Single entry configuring the threshold for free shipping to the specified zone.

## Example Usage

```hcl
resource "commercetools_shipping_method" "standard" {
  name = "Standard tax category"
  key = "Standard tax category"
  description = "Standard tax category"
  is_default = true
  tax_category_id = "<some tax category id>"
  predicate = "1 = 1"
}

resource "commercetools_shipping_zone" "de" {
  name = "DE"
  description = "Germany"
  location = {
      country = "DE"
  }
}

resource "commercetools_shipping_zone_rate" "standard-de" {
  shipping_method_id = "${commercetools_shipping_method.standard.id}"
  shipping_zone_id   = "${commercetools_shipping_zone.de.id}"

  price {
    cent_amount   = 5000
    currency_code = "EUR"
  }

  free_above {
    cent_amount   = 50000
    currency_code = "EUR"
  }
}
```

[commercetool-shipping-methods]: https://docs.commercetools.com/http-api-projects-shippingMethods.html
[commercetool-shipping-zone-rate]: https://docs.commercetools.com/http-api-projects-shippingMethods.html#shippingrate
