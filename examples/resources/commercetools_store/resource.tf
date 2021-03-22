resource "commercetools_store" "standard" {
  name = {
      nl-NL = "My standard store"
  }
  key = "standard-store"

  // optional
  languages            = ["nl-NL"]
  distribution_channels = ["NL-DIST"]
  supply_channels = ["NL-SUP"]
}
