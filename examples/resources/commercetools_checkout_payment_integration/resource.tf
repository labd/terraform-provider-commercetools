resource "commercetools_checkout_payment_integration" "card" {
  key                   = "adyen-card"
  application_id        = commercetools_checkout_application.shop.id
  connect_deployment_id = commercetools_connect_deployment.adyen.id
  name                  = "Adyen card"
  type                  = "card"
  component_type        = "Component"
  status                = "Active"
  predicate             = "country='NL' or country='BE'"
  sort_priority         = 1

  label = {
    en = "Credit card"
    nl = "Creditcard"
  }

  automated_reversal = "Active"
}
