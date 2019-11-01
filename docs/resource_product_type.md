# Product Types

Product types are used to describe common characteristics, most importantly common custom attributes, of many concrete products.

Please note: to customize other resources than products, please [refer to types](resource_type.md).

Also see the [product type HTTP API documentation][commercetool-product-type].

## Example Usage

```hcl
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
```

## Argument Reference

The following arguments are supported:

* `key` - The unique key of the product type.
* `name` - The name of the product type.
* `description` - The description of the product type.
* `attribute` - Can be 1 or more [attribute definitions](#attribute-definition)

### Attribute Definition
[Attribute Definitions][commercetool-attribute-definition] describe custom attributes and allow you to define some meta-information associated with the attribute.

These can have the following arguments:

* `type` - The type of the attribute as [Attribute Type](#attribute-type)
* `name` - The name of the attribute.<br>
    The name must be between two and 36 characters long and can contain the ASCII letters A to Z in lowercase or uppercase, digits, underscores (_) and the hyphen-minus (-).
    When using the same name for an attribute in two or more product types all fields of the AttributeDefinition of this attribute need to be the same across the product types, otherwise an AttributeDefinitionAlreadyExists error code will be returned. An exception to this are the values of an enum or lenum type and sets thereof.
* `label` - A human-readable label for the attribute as [localized string](#localized-string).
* `required` - (Optional) Whether the attribute is required to have a value.
* `input_hint` - (Optional) Provides a visual representation type for this attribute. It is only relevant for string-based attribute types like 'text' and 'ltext'. Must be one of:
    - SingleLine
    - MultiLine
* `input_tip` - (Optional) Additional information about the attribute that aids content managers when setting product details.
* `constraint` - (Optional) Describes how an attribute or a set of attributes should be validated across all variants of a product. Must be one of:
    - None (No constraints are applied to the attribute)
    - Unique (Attribute value should be different in each variant)
    - CombinationUnique (A set of attributes, that have this constraint, should have different combinations in each variant)
    - SameForAll (Attribute value should be the same in all variants)
* `searchable` - (Optional) Whether the attribute’s values should generally be enabled in product search. <br>
    This determines whether the value is stored in products for matching terms in the context of full-text search queries and can be used in facets & filters as part of product search queries. 
    The exact features that are enabled/disabled with this flag depend on the concrete attribute type and are described there. 
    The max size of a searchable field is restricted to 10922 characters. This constraint is enforced at both product creation and product update. If the length of the input exceeds the maximum size an InvalidField error is returned.

### Attribute Type
Describes the type of the field.

These can have the following arguments:

* `name` - The name of the field type. Must be one of:
    - boolean
    - text
    - ltext
    - enum
    - lenum
    - number
    - money
    - date
    - time
    - datetime
    - reference
    - set
* `values` - (**enum** type only) The enum values, defined as an object:

        values = {
            dog = "Dog"
            cat = "Cat"
        }

* `localized_value` - (**lenum** type only) One or more Localized Value objects.
* `reference_type_id` - (**reference** type only) The name of the resource type that the value should reference. Supported values for **reference** are:
    - product
    - product-type
    - channel
    - customer
    - state
    - zone
    - shipping-method
    - category
    - review
    - key-value-document
* `type_reference` - (**nested** type only) The id of the custom product type resource you want to reference.
* `element_type` - (**set** type only) Another [Attribute Type](#attribute-type) definition that is used for the set.

### Localized String
A [Localized String][commercetool-localized-string] is used to provide a string value in multiple languages.

The way to define this in the template is as:

```hcl
value = {
    en = "Our new shiny value"
    nl = "Onze versie nieuwe waarde"
}
```

### Localized Enum
A [Localized Enum][commercetool-localized-enum] is used to provide a Enum value in multiple languages.

The way to define this in the template is as:

```hcl
localized_value {
    key = "phone"
    label = {
        en = "Phone"
        nl = "Telefoon"
    }
}
```

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the Product Type.
* `version` - The version of the Product Type

[commercetool-product-type]: https://docs.commercetools.com/http-api-projects-productTypes.html
[commercetool-localized-string]: https://docs.commercetools.com/http-api-types.html#localizedstring
[commercetool-attribute-definition]: https://docs.commercetools.com/http-api-projects-productTypes.html#attributedefinition
[commercetool-localized-enum]: https://docs.commercetools.com/http-api-projects-productTypes.html#localizedenumvalue
