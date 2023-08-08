# With target lineItems and discount value relative
resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my-cart-discount-key"
  name = {
    en = "My Discount name"
  }
  description = {
    en = "My Discount description"
  }
  value {
    type      = "relative"
    permyriad = 1000
  }
  predicate = "1=1"
  target {
    type      = "lineItems"
    predicate = "1=1"
  }
  sort_order             = "0.9"
  is_active              = true
  valid_from             = "2020-01-02T15:04:05.000Z"
  valid_until            = "2021-01-02T15:04:05.000Z"
  requires_discount_code = true
  stacking_mode          = "Stacking"
}

# With target customLineItems and discount value fixed
resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my-cart-discount-key"
  name = {
    en = "My Discount name"
  }
  description = {
    en = "My Discount description"
  }
  value {
    type = "fixed"
    money {
      currency_code = "USD"
      cent_amount   = "3000"
    }
    money {
      currency_code = "EUR"
      cent_amount   = "4000"
    }
  }
  predicate = "1=1"
  target {
    type      = "customLineItems"
    predicate = "1=1"
  }
  sort_order             = "0.9"
  is_active              = true
  valid_from             = "2020-01-02T15:04:05.000Z"
  valid_until            = "2021-01-02T15:04:05.000Z"
  requires_discount_code = true
  stacking_mode          = "Stacking"
}

# With target shipping and discount value absolute
resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my-cart-discount-key"
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
      cent_amount   = "3000"
    }
    money {
      currency_code = "EUR"
      cent_amount   = "4000"
    }
  }
  predicate = "any-predicate"
  target {
    type = "shipping"
  }
  sort_order             = "0.8"
  is_active              = false
  requires_discount_code = false
  stacking_mode          = "StopAfterThisDiscount"
}

# With target multiBuyLineItems and discount value relative
resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my-cart-discount-key"
  name = {
    en = "My Discount name"
  }

  value {
    type      = "relative"
    permyriad = 1000
  }
  predicate = "any-predicate"
  target {
    type                = "multiBuyLineItems"
    predicate           = "1=1"
    trigger_quantity    = "2"
    discounted_quantity = "1"
    max_occurrence      = "1"
    selection_mode      = "MostExpensive"
  }
  sort_order = "0.8"
}

# With target multiBuyCustomLineItems and discount value relative
resource "commercetools_cart_discount" "my-cart-discount" {
  key = "my-cart-discount-key"
  name = {
    en = "My Discount name"
  }

  value {
    type      = "relative"
    permyriad = 1000
  }
  predicate = "any-predicate"
  target {
    type                = "multiBuyCustomLineItems"
    predicate           = "1=1"
    trigger_quantity    = "2"
    discounted_quantity = "1"
    max_occurrence      = "1"
    selection_mode      = "MostExpensive"
  }
  sort_order = "0.8"
}
