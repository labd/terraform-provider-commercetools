# Cart Discounts

Cart discounts are used to change the prices of different elements within a cart.

Also see the [Product Discounts HTTP API documentation](https://docs.commercetools.com/http-api-projects-productDiscounts).

## Example Usage

```hcl
resource "commercetools_product_discount" "my_absolute_product_discount" {
  name        = {
    en = "My absolute product discount name"
  }
  key         = "my-absolute-product-discount-key"
  description = {
    en = "My absolute product discount description"
  }
  predicate   = "1=1"
  sort_order  = "0.1"
  is_active   = true
  valid_from  = "2021-01-01T00:00:00.000Z"
  valid_until = "2022-01-01T00:00:00.000Z"
  value {
    type = "absolute"
    money {
      currency_code = "EUR"
      cent_amount   = 1
    }
    money {
      currency_code = "CHF"
      cent_amount   = 1
    }
  }
}

resource "commercetools_product_discount" "my_external_product_discount" {
  name        = {
    en = "My external product discount name"
  }
  key         = "my-external-product-discount-key"
  description = {
    en = "My external product discount description"
  }
  predicate   = "1=1"
  sort_order  = "0.2"
  is_active   = true
  valid_from  = "2021-01-01T00:00:00.000Z"
  valid_until = "2022-01-01T00:00:00.000Z"
  value {
    type = "external"
  }
}

resource "commercetools_product_discount" "my_relative_product_discount" {
  name        = {
    en = "My relative product discount name"
  }
  key         = "my-relative-product-discount-key"
  description = {
    en = "My relative product discount description"
  }
  predicate   = "1=1"
  sort_order  = "0.3"
  is_active   = true
  valid_from  = "2021-01-01T00:00:00.000Z"
  valid_until = "2022-01-01T00:00:00.000Z"
  value {
    type = "relative"
    permyriad = 1000
  }
}
```

## Argument Reference

* `key` - string - Optional
* `name` - string
* `description` - string - Optional
* `value` - should be one of [Product Discount Value](#product-discount-value)
* `predicate` - string - should be valid [Product Predicate][commercetool-product-predicate]
* `sort_order` - string - Optional - The string must contain a number between 0 and 1
* `is_active` - boolean - Optional - By default: true
* `valid_from` - string - Optional - A JSON string representation of UTC date & time in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)
* `valid_until` - string - Optional - A JSON string representation of UTC date & time in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)



### Product Discount Value
[Product Discount Value][commercetool-product-discount-value] defines the effect the discount will have.

These can have the following combination of arguments:
* `type` - string - Value: 'relative'
* `permyriad` - number - Per ten thousand. The fraction the price is reduced. 1000 will result in a 10% price reduction.
-----
* `type` - string - Value: 'absolute'
* `money` - array of [Money][commercetool-money] - The array contains money values in different currencies.
-----
* `type` - string - Value: 'external'


[commercetool-product-discount-value]: https://docs.commercetools.com/http-api-projects-productDiscounts.html#productdiscountvalue
[commercetool-product-predicate]: https://docs.commercetools.com/http-api-projects-predicates#productdiscount-predicates
[commercetool-money]: https://docs.commercetools.com/http-api-types.html#money