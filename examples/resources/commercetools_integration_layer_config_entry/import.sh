# A config entry is imported using its key.
#
# A secret value is not returned by the API. After an import the stored value
# is kept until secret_value is applied.
terraform import commercetools_integration_layer_config_entry.contentful_token CONTENTFUL_ACCESS_TOKEN
