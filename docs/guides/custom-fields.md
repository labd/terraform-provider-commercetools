---
subcategory: ""
page_title: "Custom Fields"
description: |-
    Defining Custom Types and setting Custom Fields on resources
---

# Custom Fields

[Custom Fields](https://docs.commercetools.com/api/projects/custom-fields) add
fields of your own to resources such as categories, channels, stores, business
units, carts and orders. The fields are defined by a Custom Type, a
`commercetools_type`, and a resource that uses the type sets their values in
its `custom` block.

## Defining a Custom Type

A type lists the resources it can be used on in `resource_type_ids`, and a
`field` for each of its fields. `input_hint` only applies to `String` and
`LocalizedString` fields.

```hcl
resource "commercetools_type" "category" {
  key               = "category-extra"
  resource_type_ids = ["category"]

  name = {
    en = "Category extras"
  }

  field {
    name = "featured"
    label = {
      en = "Featured"
    }
    type {
      name = "Boolean"
    }
  }

  field {
    name = "position"
    label = {
      en = "Position"
    }
    type {
      name = "Number"
    }
  }

  field {
    name = "intro"
    label = {
      en = "Introduction"
    }
    input_hint = "MultiLine"
    type {
      name = "LocalizedString"
    }
  }

  field {
    name = "layout"
    label = {
      en = "Layout"
    }
    type {
      name = "Enum"
      value {
        key   = "grid"
        label = "Grid"
      }
      value {
        key   = "list"
        label = "List"
      }
    }
  }

  field {
    name = "tags"
    label = {
      en = "Tags"
    }
    type {
      name = "Set"
      element_type {
        name = "String"
      }
    }
  }

  field {
    name = "supply_channel"
    label = {
      en = "Supply channel"
    }
    type {
      name              = "Reference"
      reference_type_id = "channel"
    }
  }

  field {
    name = "minimum_order"
    label = {
      en = "Minimum order value"
    }
    type {
      name = "Money"
    }
  }
}
```

A field's type cannot be changed once it exists: remove the field and add it
again under a new name.

## Setting Custom Fields

A resource sets its Custom Fields in its `custom` block, with the type's id and
the values in `fields`. Each value is a string with the value as JSON, so
encode it with `jsonencode`, whatever the field's type:

```hcl
resource "commercetools_category" "chairs" {
  key = "chairs"
  name = {
    en = "Chairs"
  }
  slug = {
    en = "chairs"
  }

  custom {
    type_id = commercetools_type.category.id
    fields = {
      featured = jsonencode(true)
      position = jsonencode(3)
      intro = jsonencode({
        en = "Chairs for every room."
      })
      # An enum value is its key.
      layout = jsonencode("grid")
      tags   = jsonencode(["indoor", "wood"])
      supply_channel = jsonencode({
        typeId = "channel"
        id     = commercetools_channel.warehouse.id
      })
      minimum_order = jsonencode({
        centAmount   = 5000
        currencyCode = "EUR"
      })
    }
  }
}
```

## Hiding Custom Fields from storefronts

On the [Commerce Integration Layer](commerce-integration-layer), Custom Fields
are part of the storefront's GraphQL schema. Hide the ones a storefront must not
see with `commercetools_integration_layer_field_visibility`, by the type's key:

```hcl
resource "commercetools_integration_layer_field_visibility" "visibility" {
  hidden_field {
    space      = "customField"
    type_key   = commercetools_type.category.key
    field_name = "position"
  }
}
```
