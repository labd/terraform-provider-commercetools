resource "commercetools_product_discount" "my-product-discount" {
  key = "my-product-discount-key"
  name = {
    en = "Product discount name"
  }
  description = {
    en = "Product discount description"
    }
  predicate              = "1=1"
  sort_order             = "0.9"
  is_active              = true
  valid_from             = "2018-01-02T15:04:05Z"
  valid_until            = "2019-01-02T15:04:05Z"

  value {
    type      = "relative"
    permyriad = 1000
  }
}

resource "commercetools_product_discount" "my-product-discount-absolute" {
  key = "my-product-discount-absolute-key"
  name = {
    en = "Product discount name"
  }
  description = {
    en = "Product discount description"
    }
  predicate              = "1=1"
  sort_order             = "0.9"
  is_active              = true
  valid_from             = "2018-01-02T15:04:05Z"
  valid_until            = "2019-01-02T15:04:05Z"

  value {
    type      = "absolute"
    money {
      currency_code = "EUR"
      cent_amount   = 500
    }
  }
}
