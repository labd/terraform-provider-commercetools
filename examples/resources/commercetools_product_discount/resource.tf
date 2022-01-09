resource "commercetools_product_discount" "my_product_discount" {
  name        = {
    en = "My product discount name"
  }
  key         = "my-product-discount-key"
  description = {
    en = "My product discount description"
  }
  predicate   = "1=1"
  sort_order  = "0.2"
  is_active   = false
  valid_from  = "2021-01-01T00:00:00.000Z"
  valid_until = "2022-01-01T00:00:00.000Z"
  value {
    type = "absolute"
    money {
      currency_code = "EUR"
      cent_amount   = 50
    }
    money {
      currency_code = "CHF"
      cent_amount   = 1
    }
  }
}