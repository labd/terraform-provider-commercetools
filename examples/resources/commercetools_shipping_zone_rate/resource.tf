resource "commercetools_shipping_method" "standard" {
  name = "Standard tax category"
  key = "Standard tax category"
  description = "Standard tax category"
  is_default = true
  tax_category_id = "<some tax category id>"
  predicate = "1 = 1"
}

resource "commercetools_shipping_zone" "de" {
  name = "DE"
  description = "Germany"
  location = {
      country = "DE"
  }
}

resource "commercetools_shipping_zone_rate" "standard-de" {
  shipping_method_id = "${commercetools_shipping_method.standard.id}"
  shipping_zone_id   = "${commercetools_shipping_zone.de.id}"

  price {
    cent_amount   = 5000
    currency_code = "EUR"
  }

  free_above {
    cent_amount   = 50000
    currency_code = "EUR"
  }

  shipping_rate_price_tier {
    type                = "CartScore"
    score               = 10

    price {
      cent_amount      = 5000
      currency_code    = "%[3]s"
    }
  }

  shipping_rate_price_tier {
    type                = "CartScore"
    score               = 20

    price {
      cent_amount      = 2000
      currency_code    = "%[3]s"
    }
  }
}