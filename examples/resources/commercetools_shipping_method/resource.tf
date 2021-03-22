resource "commercetools_shipping_method" "standard" {
  name = "Standard tax category"
  key = "Standard tax category"
  description = "Standard tax category"
  is_default = true
  tax_category_id = "<some tax category id>"
  predicate = "1 = 1"
}
