---
subcategory: ""
page_title: "Custom Fields"
---

## Creating a custom type

To extend the commercetools platform with custom fields, you can create custom
types that can be attached to resources. The following example shows how to
create a custom type that can be attached to categories.

For more information see
the [commercetools documentation](https://docs.commercetools.com/api/projects/types).

```hcl
resource "commercetools_type" "my-category" {
  key = "my-category"

  resource_type_ids = ["category"]

  name = {
    en = "myCategory"
  }

  description = {
    en = "My Category"
  }

  field {
    name  = "myBoolean"
    label = {
      en = "myBoolean"
    }
    required = false
    type {
      name = "Boolean"
    }
    input_hint = "SingleLine"
  }

  field {
    name  = "myNumber"
    label = {
      en = "myNumber"
    }
    required = false
    type {
      name = "Number"
    }
    input_hint = "SingleLine"
  }

  field {
    name = "mySet"

    label = {
      en = "mySet"
    }

    type {
      name = "Set"
      element_type {
        name = "String"
      }
    }
  }

  field {
    name  = "myLocalizedString"
    label = {
      en = "myLocalizedString"
    }
    required = false
    type {
      name = "LocalizedString"
    }
    input_hint = "SingleLine"
  }

  field {
    name  = "mySetOfLocalizedStrings"
    label = {
      en = "mySetOfLocalizedStrings"
    }
    required = false
    type {
      name = "Set"
      element_type {
        name = "LocalizedString"
      }
    }
    input_hint = "SingleLine"
  }
}
```

## Setting custom fields

Due to constrains within the old terraform provider framework, the custom fields
are set in terraform state as a map of strings. This means that the actual
values need to be serialized to JSON before they can be set. The following
example shows how to set the custom fields for a category with the custom type
given above.

```hcl
resource "commercetools_category" "my-category" {
  name = {
    en = "My category"
  }
  slug = {
    en = "my-category"
  }

  custom {
    type_id = commercetools_type.my-category.id
    fields  = {
      myBoolean         = jsonencode(true)
      myNumber          = jsonencode(123)
      mySet             = jsonencode(["a", "b", "c"])
      myLocalizedString = jsonencode({
        en = "English"
      })
      mySetOfLocalizedStrings = jsonencode([
        { en = "English 1" },
        { en = "English 2" }
      ])
    }
  }
}
```
