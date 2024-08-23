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
}

resource "commercetools_business_unit_division" "my-division" {
  key                = "my-division"
  name               = "My division"
  contact_email      = "my-division@my-company.com"
  store_mode         = "Explicit"
  status             = "Active"
  associate_mode     = "Explicit"
  approval_rule_mode = "Explicit"

  parent_unit {
    key = commercetools_business_unit_company.my-company.key
  }

  store {
    key = commercetools_store.my-store.key
  }

  address {
    key                    = "my-div-address-1"
    country                = "NL"
    state                  = "Utrecht"
    city                   = "Utrecht"
    street_name            = "Oudegracht"
    street_number          = "1"
    postal_code            = "3511 AA"
    additional_street_info = "Main floor"
  }

  address {
    key           = "my-div-address-2"
    country       = "NL"
    state         = "Zuid-Holland"
    city          = "Leiden"
    street_name   = "Breestraat"
    street_number = "1"
    postal_code   = "2311 CH"
  }
  billing_address_keys         = ["my-div-address-1"]
  shipping_address_keys        = ["my-div-address-1", "my-div-address-2"]
  default_billing_address_key  = "my-div-address-1"
  default_shipping_address_key = "my-div-address-1"
}
