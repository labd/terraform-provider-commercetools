# The IdP login is a singleton per Project, so it is imported using the
# Project key.
#
# The client secret is not returned by the API. After an import the stored
# secret is kept until client_secret is set and applied.
terraform import commercetools_integration_layer_idp_login.login my-project-key
