resource "commercetools_api_extension" "my-extension" {
  key = "test-case"

  destination = {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions          = ["Create", "Update"]
  }
}
