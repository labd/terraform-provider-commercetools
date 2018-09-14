# Commercetools Terraform provider
This is the Terraform provider for commercetools. It allows you to configure your
commercetools project with infrastructure-as-code principles.


## Using the provider
Setting up the commercetools credentials The provider reads the environment
variables `CTP_PROJECT_KEY`, `CTP_CLIENT_SECRET`, `CTP_CLIENT_ID`,
`CTP_AUTH_URL`, `CTP_API_URL` and `CTP_SCOPES`. This is compatible with the
"Environment Variables" format you can download in the Merchant Center after
creating an API Client.

Alternatively, you can set it up directly in the terraform file:

```hcl
provider "commercetools" {
  client_id     = "<your client id>"
  client_secret = "<your client secret>"
  project_key   = "<your project key>"
}
```


## Authors
This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-commercetools)
for more information.
