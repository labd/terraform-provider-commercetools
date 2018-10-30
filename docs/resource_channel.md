# Channel Settings

Lets you manage channels within a commercetools project.

## Example Usage

```hcl
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
```

## Argument Reference

* `key` - string - Required
* `roles` - Set of ChannelRole values - Optional
If not specified, then channel will get InventorySupply role by default
* `name` - LocalizedString - Optional
* `description` - LocalizedString - Optional
