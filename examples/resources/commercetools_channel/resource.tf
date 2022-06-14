resource "commercetools_channel" "my-channel" {
  key = "my-channel-key"
  roles = ["ProductDistribution"]
  name = {
      nl-NL = "Channel"
  }
  description = {
      nl-NL = "Channel"
  }
}
