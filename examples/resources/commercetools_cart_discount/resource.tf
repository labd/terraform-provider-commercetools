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
