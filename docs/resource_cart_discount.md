# Cart Discounts

Cart discounts are used to change the prices of different elements within a cart.

Also see the [Cart Discounts HTTP API documentation](https://docs.commercetools.com/http-api-projects-cartDiscounts).

## Example Usage

```hcl
resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my_discount"
  name = {
    en = "My Discount name"
  }
  description = {
    en = "My Discount description"
  }
  value {
    type = "relative"
    permyriad = 1000
  }
  predicate = "1=1"
  target = {
    type = "lineItems"
    predicate = "1=1"
  }
  sort_order = "0.9"
  is_active = true
  valid_from = "2020-01-02T15:04:05.000Z"
  valid_until = "2021-01-02T15:04:05.000Z"
  requires_discount_code = true
  stacking_mode = "Stacking"
}

resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my_discount"
  name = {
    en = "My Discount name"
  }
  description = {
    en = "My Discount description"
  }
  value {
    type = "absolute"
    money {
      currency_code = "USD"
      cent_amount = "3000"
    }
    money {
    currency_code = "EUR"
    cent_amount = "4000"
    }
  }
  predicate = "any-predicate"
  target = {
    type = "shipping"
  }
  sort_order = "0.8"
  is_active = false
  requires_discount_code = false
  stacking_mode = "StopAfterThisDiscount"
}

resource "commercetools_cart_discount" "my-cart-discount" {
  name = {
    en = "My Discount name"
  }
  value {
    type = "giftLineItem"
    product_id = "product-id"
    variant = 1
    supply_channel_id = "supply-channel-id"
    distribution_channel_id	= "distribution-channel-id"
  }
  predicate = "any-predicate"
  target = {
    type = "shipping"
  }
  sort_order = "0.8"
}
```

## Argument Reference

* `key` - string - Optional
* `name` - string
* `description` - string - Optional
* `value` - should be one of [Cart Discount Value](#cart-discount-value)
* `predicate` - string - should be valid [Cart Predicate][commercetool-cart-predicate]
* `target` -  should be one of [Cart Discount Target](#cart-discount-target) - Optional - Must not be set when the `value` has type 'giftLineItem', otherwise a Cart Discount Target must be set.
* `sort_order` - string - Optional - The string must contain a number between 0 and 1
* `is_active` - boolean - Optional - By default: true
* `valid_from` - string - Optional - A JSON string representation of UTC date & time in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)
* `valid_until` - string - Optional - A JSON string representation of UTC date & time in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)
* `requires_discount_code` - boolean - Optional - By default: false
* `stacking_mode` - string - Optional - should be valid [Stacking Mode][commercetool-stacking-mode]. By default: 'Stacking'



### Cart Discount Value
[Cart Discount Value][commercetool-cart-discount-value] defines the effect the discount will have.

These can have the following combination of arguments:
* `type` - string - Value: 'relative'
* `permyriad` - number - Per ten thousand. The fraction the price is reduced. 1000 will result in a 10% price reduction.
-----
* `type` - string - Value: 'absolute'
* `money` - array of [Money][commercetool-money] - The array contains money values in different currencies.
-----
* `type` - string - Value: 'giftLineItem'
* `product` - string - ID of appropriate [Product][commercetool-product]
* `variantId` - number - Number of the product's variant
* `supplyChannel` - string - Optional - Id of the [Channel][commercetool-channel]. Must have the role 'InventorySupply'
* `distributionChannel` - string - Optional - Id of the [Channel][commercetool-channel]. Must have the role 'ProductDistribution'


### Cart Discount Target
[Cart Discount Target][commercetool-cart-discount-target] defines what part of the cart will be discounted.

These can have the following combination arguments:

* `type` - string - Value: 'lineItems'
* `predicate` - string - should be valid [Line Item Predicate][commercetool-line-item-predicate]
------
* `type` - string - Value: 'customLineItems'
* `predicate` - string - should be valid [Custom Line Item Predicate][commercetool-custom-line-item-predicate]
------
* `type` - string - Value: 'shipping'



[commercetool-cart-discount-value]: https://docs.commercetools.com/http-api-projects-cartDiscounts.html#cartdiscountvalue
[commercetool-cart-predicate]: https://docs.commercetools.com/http-api-projects-predicates#cart-predicates
[commercetool-cart-discount-target]: https://docs.commercetools.com/http-api-projects-cartDiscounts#cartdiscounttarget
[commercetool-stacking-mode]: https://docs.commercetools.com/http-api-projects-cartDiscounts#stackingmode
[commercetool-money]: https://docs.commercetools.com/http-api-types.html#money
[commercetool-channel]: https://docs.commercetools.com/http-api-projects-channels.html#channels
[commercetool-product]: https://docs.commercetools.com/http-api-projects-products.html
[commercetool-line-item-predicate]: https://docs.commercetools.com/http-api-projects-predicates.html#lineitem-field-identifiers
[commercetool-custom-line-item-predicate]: https://docs.commercetools.com/http-api-projects-predicates.html#customlineitem-field-identifiers