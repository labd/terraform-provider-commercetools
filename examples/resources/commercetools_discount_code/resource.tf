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
