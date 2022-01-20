resource "commercetools_shipping_method" "standard" {
  name = "Standard shipping method"
  key = "Standard"
  description = "Standard shipping method"
  is_default = true
  tax_category_id = "<some tax category id>"
  predicate = "1 = 1"
}
