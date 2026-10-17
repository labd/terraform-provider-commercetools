resource "commercetools_integration_layer_checkout_rules" "rules" {
  rule {
    countries       = ["DE", "AT"]
    application_key = "eu-checkout"
    mode            = "COMPLETE"
  }

  rule {
    stores          = ["vip-store"]
    application_key = "vip-checkout"
    mode            = "PAYMENT_ONLY"
  }

  # Matches every other cart, so it has to be last.
  rule {
    application_key = "default-checkout"
    mode            = "PAYMENT_ONLY"
  }
}
