resource "commercetools_channel" "project" {
  key = "My channel"
  roles = ["ProductDistribution"]
  name = {
      nl-NL = "Channel"
  }
  description = {
      nl-NL = "Channel"
  }
}
