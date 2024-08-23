resource "commercetools_store" "my-store" {
  key = "my-store"
  name = {
    en-US = "My store"
  }
  countries = ["NL", "BE"]
  languages = ["en-GB"]
}

resource "commercetools_business_unit_company" "my-company" {
  key           = "my-company"
  name          = "My company"
  contact_email = "main@my-company.com"

  address {
    key                    = "my-company-address-1"
    country                = "NL"
    state                  = "Noord-Holland"
    city                   = "Amsterdam"
    street_name            = "Keizersgracht"
    street_number          = "3"
    additional_street_info = "4th floor"
    postal_code            = "1015 CJ"
  }

  address {
    key                    = "my-company-address-2"
    country                = "NL"
    state                  = "Utrecht"
    city                   = "Utrecht"
    street_name            = "Oudegracht"
    street_number          = "1"
    postal_code            = "3511 AA"
    additional_street_info = "Main floor"
  }

  store {
    key = commercetools_store.my-store.key
  }

  billing_address_keys         = ["my-company-address-1"]
  shipping_address_keys        = ["my-company-address-1", "my-company-address-2"]
  default_billing_address_key  = "my-company-address-1"
  default_shipping_address_key = "my-company-address-1"
}
