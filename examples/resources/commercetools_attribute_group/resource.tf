resource "commercetools_attribute_group" "my-attribute-group" {
  key = "my-attribute-group-key"
  name = {
    en = "my-attribute-group-name"
  }
  description = {
    en = "my-attribute-group-description"
  }

  attribute {
    key = "attribute-key-1"
  }

  attribute {
    key = "attribute-key-2"
  }
}
