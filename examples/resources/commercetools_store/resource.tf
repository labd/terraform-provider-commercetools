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

resource "commercetools_store" "my-store" {
  key = "my-store"
  name = {
    en-US = "My store"
  }
  languages             = ["en-US"]
  distribution_channels = ["US-DIST"]
  supply_channels       = ["US-SUP"]

  depends_on = [
    commercetools_channel.us-supply-channel,
    commercetools_channel.us-dist-channel,
  ]
}
