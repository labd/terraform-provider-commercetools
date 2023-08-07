resource "commercetools_type" "my-custom-type" {
  key = "my-custom-type"
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
  }

  field {
    name = "contact_time"
    label = {
      en = "Contact time"
      nl = "Contact tijd"
    }
    type {
      name = "Enum"
      value {
        key   = "day"
        label = "Daytime"
      }
      value {
        key   = "evening"
        label = "Evening"
      }
    }
  }

  field {
    name = "emails"

    label = {
      en = "Emails"
      nl = "Emails"
    }

    type {
      name = "Set"
      element_type {
        name = "String"
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
        label = {
          en = "Phone"
          nl = "Telefoon"
        }
      }
      localized_value {
        key = "skype"
        label = {
          en = "Skype"
          nl = "Skype"
        }
      }
    }
  }
}
