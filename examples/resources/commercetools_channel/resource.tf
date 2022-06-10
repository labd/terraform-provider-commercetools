resource "commercetools_channel" "my-channel" {
  key = "My channel"
  roles = ["ProductDistribution"]
  name = {
      nl-NL = "Channel"
  }
  description = {
      nl-NL = "Channel"
  }
}
