# Discount Codes

With discount codes it is possible to give specific cart discounts to an eligible set of users. They are defined by a string value which can be added to a cart so that specific cart discounts can be applied to the cart.


Also see the [Discount Codes HTTP API documentation](https://docs.commercetools.com/http-api-projects-discountCodes#discount-codes).

## Example Usage

```hcl
resource "commercetools_discount_code" "my_discount_code" {
  name = {
    en = "My Discount code name"
  }
  description = {
    en = "My Discount code description"
  }
  code = "2"
  valid_from = "2020-01-02T15:04:05.000Z"
  valid_until = "2021-01-02T15:04:05.000Z"
  is_active = true
  predicate = "1=1"
  max_applications_per_customer = 3
  max_applications = 100
  groups = ["0", "1"]
  cart_discounts = ["cart-discount-id-1", "cart-discount-id-2"]
}

resource "commercetools_discount_code" "my_discount_code" {
  code = "2"
  cart_discounts = ["cart-discount-id-1"]
}
```

## Argument Reference

* `name` - string - Optional
* `description` - string - Optional
* `code` - string
* `valid_from` - string - Optional - A JSON string representation of UTC date & time in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)
* `valid_until` - string - Optional - A JSON string representation of UTC date & time in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)
* `is_active` - boolean - Optional - By default: true
* `predicate` - string - should be valid [Cart Predicate][commercetool-cart-predicate]
* `max_applications_per_customer` - number - Optional - The discount code can only be applied `max_applications_per_customer` times per customer.
* `max_applications` - number - Optional - The discount code can only be applied `max_applications` times.
* `groups` - []string - Optional - The groups to which this discount code belong.
* `cart_discounts` - []string - The array of [Cart Discounts][commercetool-cart-discount] IDs



[commercetool-cart-predicate]: https://docs.commercetools.com/http-api-projects-predicates#cart-predicates
[commercetool-cart-discount]: https://docs.commercetools.com/http-api-projects-cartDiscounts.html#cartdiscount
