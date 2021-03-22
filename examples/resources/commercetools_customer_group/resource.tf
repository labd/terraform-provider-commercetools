resource "commercetools_customer_group" "standard" {
  name = "Standard Customer Group"
  key  = "standard-customer-group"
}

resource "commercetools_customer_group" "golden" {
  name = "Golden Customer Group"
  key  = "golden-customer-group"
}
