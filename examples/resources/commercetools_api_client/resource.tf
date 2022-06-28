resource "commercetools_api_client" "my-api-client" {
  name  = "My API Client"
  scope = ["manage_orders:my-ct-project-key", "manage_payments:my-ct-project-key"]
}
