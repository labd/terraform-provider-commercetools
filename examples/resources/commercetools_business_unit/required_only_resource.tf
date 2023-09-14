resource "commercetools_business_unit" "acme_business_unit" {
    key = "acme-business-unit"
    status = "Active"
    store_mode = "Explicit"
    unit_type = "Company"

    stores {
        key = "acme-store-dusseldorf"
        type_id = "store"
    }

    stores {
        key = "acme-store-berlin"
        type_id = "store"
    }

    addresses {
        key = "acme-business-unit-address"
        title = "Acme Business Unit Address"
        salutation = "Mr."
        first_name = "John"
        last_name = "Doe"
        street_name = "Main Street"
        street_number = "1"
        additional_street_info = "Additional Street Info"
        postal_code = "12345"
        city = "Berlin"
        region = "Berlin"
        country = "DE"
        company = "Acme"
        department = "IT"
        building = "Building"
        apartment = "Apartment"
        p_o_box = "P.O. Box"
        phone = "123456789"
        mobile = "987654321"
    }

    default_shipping_address_id = "acme-business-unit-address"
    default_billing_address_id = "acme-business-unit-address"
}
