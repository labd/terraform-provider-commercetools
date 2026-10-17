resource "commercetools_integration_layer_field_visibility" "visibility" {
  hidden_field {
    space      = "customField"
    type_key   = "order-extra"
    field_name = "costPrice"
  }

  hidden_field {
    space      = "attribute"
    type_key   = "chair"
    field_name = "supplierCode"
  }
}
