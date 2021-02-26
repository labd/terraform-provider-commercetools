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
