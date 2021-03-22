resource "commercetools_product_type" "some-generic-properties-product-type" {
    name = "Some generic product properties"
    description = "All the generic product properties"

    attribute {
        name = "perishable"
        label = {
            en = "Is perishable"
            nl = "Is perishable"
        }
        required = true
        type {
            name = "boolean"
        }
    }
}

resource "commercetools_product_type" "my-product-type" {
    name = "Lens specification"
    description = "All the specific info concerning the lens"

    attribute {
        name = "autofocus"
        label = {
            en = "Has autofocus"
            nl = "Heeft autofocus"
        }
        required = true
        type {
            name = "boolean"
        }
    }

    attribute {
        name = "lens_product_no"
        label = {
            en = "Lens product number"
            nl = "Objectief productnummer"
        }
        required = true
        type {
            name = "text"
        }
        constraint = "Unique"
        input_tip = {
            en = "Enter the product code"
            nl = "Voer de product code in"
        }
        searchable = true
    }

    attribute {
        name = "previous_model"
        label = {
            en = "Previous model"
            nl = "Vorig model"
        }
        type = {
            name = "reference"
            reference_type_id = "product"
        }
    }

    attribute {
        name = "product_properties"
        label = {
            en = "Product properties"
            nl = "Product eigenschappen"
        }
        required = false
        type {
            name =  "nested"
            type_reference = "${commercetools_product_type.some-generic-properties-product-type.id}"
        }
    }
}
