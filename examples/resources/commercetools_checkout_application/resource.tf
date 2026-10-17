resource "commercetools_checkout_application" "shop" {
  key                           = "shop"
  name                          = "Shop"
  mode                          = "CompleteFlow"
  status                        = "Active"
  countries                     = ["NL", "BE"]
  allowed_origins               = ["https://shop.example.com"]
  logo_url                      = "https://shop.example.com/logo.svg"
  payment_return_url            = "https://shop.example.com/checkout/return"
  active_payment_component_type = "Component"
  allow_discounts               = true

  description = {
    en = "Checkout for the shop"
  }

  agreement {
    name = "terms"
    type = "MandatoryCheckbox"
    text = {
      en = "I agree to the terms and conditions."
    }
  }
}
