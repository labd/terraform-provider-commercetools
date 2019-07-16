# Product Discounts

A product discount applies to a specific product or subset of products based criteria you provide. 
Also see [Commercetools API docs](commercetools-product-discounts).

## Example Usage

```hcl
resource "commercetools_product_discount" "10_percent_off" {
  key  = "10-percent-off"
  name = {
      nl-NL = "10% korting"
      en    = "10% discount"
  }
  description = {
      nl-NL = "10% korting"
      en    = "10% discount"
  } 
  sort_order = "0.05"
  is_active  = true
  value = {
    type      = "relative"
    permyriad = 1000
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - Localized name of the product discount.
* `sort_order` - String to order the discounts, must be between 0 and 1.
* `value` - Actual value of the discount. For more options see Product discount values.

* `key` - (Optional) User-specific unique identifier for the product discount.
* `description` - (Optional) Localized description of the product discount.
* `predicate` - (Optional) Predicate to match this product discount.
* `is_active` - (Optional) If this discount is active, default false.
* `valid_from` - (Optional) Date in format YYYY-MM-DD, when the discount should be active.
* `valid_until` - (Optional) Date in format YYYY-MM-DD, when the discount should be active.


### Product discount values
There are three types of product discount types: external, relative and absolute. Below are the arguments for each type.

#### Absolute

* `type` - 'relative'
* `currency_code` - A three-digit currency code as per ISO 4217.
* `cent_amount` - Amount in cents.

#### Relative

* `type` - 'relative'
* `permyriad` - Per ten thousand. The fraction the price is reduced. 1000 will result in a 10% price reduction. 

#### External

* `type` - 'external'

[commercetools-product-discounts](https://docs.commercetools.com/http-api-projects-productDiscounts)
