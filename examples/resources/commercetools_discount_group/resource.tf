resource "commercetools_discount_group" "summer_sale" {
  key        = "summer-sale"
  sort_order = "0.6"
  name = {
    en = "Summer sale"
  }
}

# Only the best discount of the group applies to a cart.
resource "commercetools_cart_discount" "ten_percent" {
  name = {
    en = "Ten percent off"
  }
  predicate         = "1 = 1"
  discount_group_id = commercetools_discount_group.summer_sale.id
  target {
    type      = "lineItems"
    predicate = "1 = 1"
  }
  value {
    type      = "relative"
    permyriad = 1000
  }
}
