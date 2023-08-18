data "commercetools_type" "existing_type" {
  key = "test"
}

resource "commercetools_channel" "test" {
  key   = "test"
  roles = ["ProductDistribution"]
  custom {
    type_id = data.commercetools_type.existing_type.id
    fields = {
      "my-field" = "foobar"
    }
  }
}
