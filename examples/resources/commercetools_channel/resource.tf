resource "commercetools_channel" "project" {
  key = "My project"
  roles = ["ProductDistribution"]
  name = {
      nl-NL = "Channel"
  }
  description = {
      nl-NL = "Channel"
  }
}
