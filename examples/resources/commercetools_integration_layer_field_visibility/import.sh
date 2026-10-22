# The hidden fields are a singleton per Project, so they are imported using
# the Project key. Fields hidden in the Merchant Center are imported too.
terraform import commercetools_integration_layer_field_visibility.visibility my-project-key
