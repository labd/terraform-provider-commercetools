# Custom Object

Manages a custom object value. When the `container` or `key` value is modified
it will delete the old custom object and create a new custom object instead of
updating the value.

The value is always a string, so use `jsonencode()` to convert an object to a
string.

## Example Usage

```hcl
resource "commercetools_custom_object" "my-value" {
  container = "my-container"
  key = "my-key"
  value = jsonecode(10)
}
```

## Argument Reference

The following arguments are supported:

* `container` - The container
* `key` - The key to save the value in the container
* `value` - A string (can be json)
