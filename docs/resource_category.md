# Categories

Categories allow to organize products into hiearchical structures.

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
* `key` - string - Optional
* `name` - LocalizedString - optional
* `description` - LocalizedString - Optional
* `slug` - LocalizedString - optional - human readable identifiers, needs to be unique
* `parent` - string - optional - A category that is the parent of this category in the category tree
* `order_hint` -  string - optional - An attribute as base for a custom category order in one level, filled with random value when left empty
* `external_id` - string - Optional
* `meta_title` - LocalizedString - Optional
* `meta_description` - LocalizedString - Optional
* `meta_keywords` - LocalizedString - Optional
