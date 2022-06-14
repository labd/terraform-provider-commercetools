resource "commercetools_customer_group" "standard" {
  key = "my-customer-group-key"
  name = "Standard Customer Group"
}

resource "commercetools_customer_group" "golden" {
  key = "my-customer-group-key"
  name = "Golden Customer Group"
}
