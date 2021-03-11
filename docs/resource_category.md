# Categories

Categories allow you to organize products into hierarchical structures.

Also see the [Categories HTTP API documentation](https://docs.commercetools.com/api/projects/categories).

## Example Usage

```hcl
resource "commercetools_category" "example" {
  name = {
    en = "example"
  }
  key = "example"
  description = {
    en = "Standard description"
  }
  parent = "<id-of-parent>"
  slug = {
    en = "example"
  }
  meta_title = {
    en = "Meta title"
  }
  assets {
    key = "some_key"
    name = {
      en = "Image name"
    }
    description = {
      en = "Image description"
    }
    sources {
      uri = "https://example.com/test.jpg"
      key = "image"
    }
  }
}
```

## Argument Reference
* `key` - String - Optional
* `name` - LocalizedString - Optional
* `description` - LocalizedString - Optional
* `slug` - LocalizedString - Optional - Human readable identifiers, needs to be unique
* `parent` - String - Optional - A category that is the parent of this category in the category tree
* `order_hint` -  String - Optional - An attribute as base for a custom category order in one level, filled with random value when left empty
* `external_id` - String - Optional
* `meta_title` - LocalizedString - Optional
* `meta_description` - LocalizedString - Optional
* `meta_keywords` - LocalizedString - Optional
* `assets` - Array of [Asset](#asset) - Optional

### Asset
[Asset][asset] defines an image, icon or movie that is related to this category.
* `key` - String - Human readable identifier for the asset
* `sources` - Array of [AssetSource](#asset-source)
* `description` - String - Optional
* `tags` - String Array - Optional

### AssetSource
[AssetSource][asset-source] is a representation of an asset in a specific file format.
* `uri` - String - Required
* `key` - String - Optional
* `dimensions` - Instance of [AssetDimensions](#asset-dimensions)
* `content_type` - String - Optional

### AssetDimensions
[AssetDimensions][asset-dimensions] specifies the width and height of an asset.
* `w` - String - Required
* `h` - String - Required

[asset]: https://docs.commercetools.com/api/types#asset
[asset-source]: https://docs.commercetools.com/api/types#assetsource
[asset-dimensions]: https://docs.commercetools.com/api/types#assetdimensions
