resource "commercetools_type" "my-type" {
  key = "my-type"
  name = {
    en = "My type"
    nl = "Mijn type"
  }

  resource_type_ids = ["product-selection"]

  field {
    name = "my-field"
    label = {
      en = "My field"
      nl = "Mijn veld"
    }
    type {
      name = "String"
    }
  }
}

resource "commercetools_product_selection" "product-selection-us" {
  key = "product-selection-us"
  name = {
    en = "US Product Selection"
  }
  mode = "Individual"

  custom {
    type_id = commercetools_type.my-type.id
    fields = {
      my-field = "my-value"
    }
  }
}
