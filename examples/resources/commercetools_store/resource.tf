resource "commercetools_channel" "us-supply-channel" {
  key   = "US-SUP"
  roles = ["InventorySupply"]
  name = {
    en-US = "Supply channel"
  }
  description = {
    en-US = "Supply channel desc"
  }
}

resource "commercetools_channel" "us-dist-channel" {
  key   = "US-DIST"
  roles = ["ProductDistribution"]
  name = {
    en-US = "Dist channel"
  }
  description = {
    en-US = "Dist channel desc"
  }
}

resource "commercetools_type" "my-store-type" {
  key = "my-custom-store-type"
  name = {
    en = "My Store Type"
  }
  description = {
    en = "A custom store type"
  }

  resource_type_ids = ["store"]

  field {
    name = "some-field"
    label = {
      en = "Some Field"
    }
    type {
      name = "String"
    }
  }
}


resource "commercetools_store" "my-store" {
  key = "my-store"
  name = {
    en-US = "My store"
  }
  countries             = ["NL", "BE"]
  languages             = ["en-US"]
  distribution_channels = ["US-DIST"]
  supply_channels       = ["US-SUP"]

  custom {
    type_id = commercetools_type.my-store-type.id
    fields = {
      my-field = "ja"
    }
  }

  depends_on = [
    commercetools_channel.us-supply-channel,
    commercetools_channel.us-dist-channel,
  ]
}
