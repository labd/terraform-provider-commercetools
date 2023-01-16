---
subcategory: ""
page_title: "Custom Fields"
---

## Types example

```hcl
resource "commercetools_type" "ctype1" {
  key = "contact_info"

  resource_type_ids = ["customer"]

  name = {
    en = "Contact info"
    nl = "Contact informatie"
  }

  description = {
    en = "All things related communication"
    nl = "Alle communicatie-gerelateerde zaken"
  }


  field {
    name = "skype_name"
    label = {
      en = "Skype name"
      nl = "Skype naam"
    }
    type {
      name = "String"
    }
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
