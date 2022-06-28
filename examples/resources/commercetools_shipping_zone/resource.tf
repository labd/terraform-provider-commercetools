resource "commercetools_shipping_zone" "de-us" {
  key         = "some-key"
  name        = "DE and US"
  description = "Germany and US"
  location {
    country = "DE"
  }
  location {
    country = "US"
    state   = "Nevada"
  }
}
