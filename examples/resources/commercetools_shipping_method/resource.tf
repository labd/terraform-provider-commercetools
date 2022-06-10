resource "commercetools_tax_category" "some-tax-category" {
	name = "some test cateogry"
	description = "test category"
}

resource "commercetools_shipping_method" "standard" {
  name = "Standard tax category"
  key = "Standard tax category"
  description = "Standard tax category"
  is_default = true
  tax_category_id = commercetools_tax_category.some-tax-category.id
  predicate = "1 = 1"
}
