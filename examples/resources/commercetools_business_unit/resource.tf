resource "commercetools_business_unit_company" "acme_company" {
    key              = "acme-company"
    name             = "The ACME Company"
    status           = "Active"
    contact_email    = "acme@example.com"

    store {
        key = "acme-usa"
        type_id = "store"
    }

    store {
        key = "acme-germany"
        type_id = "store"
    }

    address {
        key                     = "acme-business-unit-address"
        title                   = "Acme Business Unit Address"
        salutation              = "Mr."
        first_name              = "John"
        last_name               = "Doe"
        street_name             = "Main Street"
        street_number           = "1"
        additional_street_info  = "Additional Street Info"
        postal_code             = "12345"
        city                    = "Berlin"
        region                  = "Berlin"
        country                 = "DE"
        company                 = "Acme"
        department              = "IT"
        building                = "Building"
        apartment               = "Apartment"
        po_box                  = "P.O. Box"
        phone                   = "123456789"
        mobile                  = "987654321"
    }

    default_shipping_address_id     = "acme-business-unit-address"
    default_billing_address_id      = "acme-business-unit-address"
}

resource "commercetools_business_unit_division" "acme-willie-coyote" {
    key              = "acme-willie-coyote"
    name             = "Willie Coyote - Traps for Roadrunners"
    status           = "Active"
    contact_email    = "acme-traps@example.com"
    store_mode       = "FromParent"
    associate_mode   = "ExplicitAndFromParent"

    // Only available for division business units as the Company
    // business unit has no parent unit and must always be the Top Level Unit.
    parent_unit {
        key         = commercetools_business_unit_company.acme-company.key
        type_id     = "company"
    }
}
