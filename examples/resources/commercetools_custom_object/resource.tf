resource "commercetools_custom_object" "my-custom-object" {
  container = "my-container"
  key       = "my-key"
  value     = jsonencode(10)
}
