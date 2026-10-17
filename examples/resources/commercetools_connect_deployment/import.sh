# A deployment is imported using its key.
#
# Secured configuration values are not returned by the API. Set them in the
# configuration before the first apply after an import: a redeploy sends the
# configured values, and none when they are left out.
terraform import commercetools_connect_deployment.search_sync search-sync
