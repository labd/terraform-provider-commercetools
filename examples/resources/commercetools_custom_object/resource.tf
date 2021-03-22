resource "commercetools_custom_object" "my-value" {
  container = "my-container"
  key = "my-key"
  value = jsonencode(10)
}
