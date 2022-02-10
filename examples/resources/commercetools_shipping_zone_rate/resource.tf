resource "commercetools_tax_category" "my-tax-category" {
  key         = "some-tax-category-key"
  name        = "My tax category"
  description = "Example"
}

resource "commercetools_shipping_method" "my-shipping-method" {
  key             = "some-shipping-method-key"
  name            = "My shipping method"
  description     = "Standard method"
  is_default      = true
  tax_category_id = commercetools_tax_category.my-tax-category.id
  predicate       = "1 = 1"
}

resource "commercetools_shipping_zone" "my-shipping-zone" {
  key         = "some-shipping-zone-key"
  name        = "DE"
  description = "My shipping zone"
  location {
    country = "DE"
  }
}

resource "commercetools_shipping_zone_rate" "my-shipping-zone-rate" {
  shipping_method_id = commercetools_shipping_method.my-shipping-method.id
  shipping_zone_id   = commercetools_shipping_zone.my-shipping-zone.id

  price {
    cent_amount   = 5000
    currency_code = "EUR"
  }

  free_above {
    cent_amount   = 50000
    currency_code = "EUR"
  }

  shipping_rate_price_tier {
    type  = "CartScore"
    score = 10

    price {
      cent_amount   = 5000
      currency_code = "EUR"
    }
  }

  shipping_rate_price_tier {
    type  = "CartScore"
    score = 20

    price {
      cent_amount   = 2000
      currency_code = "EUR"
    }
  }

  shipping_rate_price_tier {
    type  = "CartScore"
    score = 30

    price_function {
      function      = "x + 1"
      currency_code = "EUR"
    }
  }
}
