resource "commercetools_tax_category" "standard" {
  name = "Standard tax category"
  key  = "standard-tax-category"
}

resource "commercetools_tax_category_rate" "standard-tax-category-DE" {
  tax_category_id   = commercetools_tax_category.my-tax-category.id
  name              = "19% MwSt"
  amount            = 0.19
  included_in_price = false
  country           = "DE"
}

resource "commercetools_tax_category_rate" "standard-tax-category-NL" {
  tax_category_id   = commercetools_tax_category.my-tax-category.id
  name              = "21% BTW"
  amount            = 0.21
  included_in_price = true
  country           = "NL"
}
