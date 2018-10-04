# Custom Types

!!! note
    Will be introduced in 0.2.0

Types define custom fields that are used to enhance resources as you need.
Since there is no model that fits all use-cases, we give you the possibility
to customize some resources, so that they fit your data model as close as
possible.

Use Types to model your own CustomFields on resources, like Category and
Customer.

In case you want to customize products, please use product types instead that
serve a similar purpose, but tailored to products.

 - product types are specialized to customize products.
 - types are used to customize other resources.


# Type

Provides a commercetools type

Also see the [type HTTP API documentation][commercetool-type].

## Example Usage

```hcl
resource "commercetools_type" "my-custom-type" {
  key = "contact_info"
  name = {
    en = "Contact info"
    nl = "Contact informatie"
  }
  description = {
    en = "All things related communication"
    nl = "Alle communicatie-gerelateerde zaken"
  }

  resource_type_ids = ["customer"]

  field {
    name = "skype_name"
    label = {
      en = "Skype name"
      nl = "Skype naam"
    }
    type {
      name = "String"
    }

  field {
    name = "contact_time"
    label = {
      en = "Contact time"
      nl = "Contact tijd"
    }
    type {
      name = "Enum"
      values {
        day = "Daytime"
        evening = "Evening"
      }
    }
  }

  field {
    name = "contact_preference"
    label = {
      en = "Contact preference"
      nl = "Contact voorkeur"
    }
    type {
      name = "LocalizedEnum"
      localized_value {
        key = "phone"
        label {
          en = "Phone"
          nl = "Telefoon"
        }
      }
      localized_value {
        key = "skype"
        label {
          en = "Skype"
          nl = "Skype"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `key` - The unique key of the Type.
* `name` - The name of the Type as [localized string](#localized-string).
* `description` - The description of the Type as [localized string](#localized-string).
* `resource_type_ids` - An array of types that can be customized with this Type.<br>
This can be any of the following:
    - asset
    - category
    - channel
    - customer
    - customer-group
    - cart-discount
    - discount-code
    - inventory-entry
    - order
    - line-item
    - custom-line-item
    - product-price
    - payment
    - payment-interface-interaction
    - shopping-list
    - shopping-list-text-line-item
    - review
* `field` - Can more 1 our more [field definitions](#field-definition) definitions

### Field Definition
[Field Definitions][commercetool-field-definition] describe custom fields and allow you to define some meta-information associated with the field.

These can have the following arguments:

* `type` - The type of the field as Field [Type](#field-type)
* `name` - The name of the field.<br>
    The name must be between two and 36 characters long and can contain the ASCII letters A to Z in lowercase or uppercase, digits, underscores (_) and the hyphen-minus (-).
* `label` - A human-readable label for the field as [localized string](#localized-string).
* `required` - (Optional) Whether the field is required to have a value.
* `input_hint` - (Optional) Provides a visual representation type for this field. It is only relevant for string-based field types like String and LocalizedString.

### Field Type
Describes the type of the field.

These can have the following arguments:

* `name` - The name of the field type. Must be one of:
    - Boolean
    - String
    - LocalizedString
    - Enum
    - LocalizedEnum
    - Number
    - Money
    - Date
    - Time
    - DateTime
    - Reference
    - Set
* `values` - (**Enum** type only) The enum values, defined as an object:

        values = {
            dog = "Dog"
            cat = "Cat"
        }

* `localized_value` - (**LocalizedEnum** type only) One or more Localized Value objects.
* `reference_type_id` - (**Reference** type only) The name of the resource type that the value should reference. Supported values are:
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

### Localized String
A [Localized String][commercetool-localized-string] is used to provide a string value in multiple languages.

The way to define this in the template is as:

```hcl
value {
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
    label {
        en = "Phone"
        nl = "Telefoon"
    }
}
```

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the Type.
* `version` - The version of the Type


[commercetool-type]: https://docs.commercetools.com/http-api-projects-types.html
[commercetool-localized-string]: https://docs.commercetools.com/http-api-types.html#localizedstring
[commercetool-field-definition]: https://docs.commercetools.com/http-api-projects-types.html#fielddefinition
[commercetool-localized-enum]: https://docs.commercetools.com/http-api-projects-types.html#localizedenumvalue
