# Zones Settings

Lets you manage shipping zones within a commercetools project.

Also see the [Zones HTTP API documentation][commercetool-zones].

## Example Usage

```hcl
resource "commercetools_shipping_zone" "de-us" {
  name = "DE and US"
  description = "Germany and US"
  location {
      country = "DE"
  }
  location {
      country = "US"
      state = "Nevada"
  }
}
```

## Argument Reference

* `name` - string
* `description` - string - Optional
* `location` - 1 or more of [Location][#location] values

### Location
[Location][commercetool-locations] defines a specific location.

These can have the following arguments:

* `country` - A two-digit country code as per ISO 3166-1 alpha-2
* `state` - string - Optional


[commercetool-zones]: https://docs.commercetools.com/http-api-projects-zones
[commercetool-locations]: https://docs.commercetools.com/http-api-projects-zones#location
