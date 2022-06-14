resource "commercetools_category" "my-category" {
  key = "my-category-key"

  name = {
    en = "My category"
  }
  description = {
    en = "Standard description"
  }
  slug = {
    en = "my_category"
  }
  meta_title = {
    en = "Meta title"
  }
}

resource "commercetools_category" "my-second-category" {
  key = "my-category-key"

  name = {
    en = "Second category"
  }
  description = {
    en = "Standard description"
  }
  parent = commercetools_category.my-category.id
  slug = {
    en = "my_second_category"
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
