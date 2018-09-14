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

## Using with docker

The included `Dockerfile` bundles the official  [`hashicorp/terraform:light`](https://hub.docker.com/r/hashicorp/terraform/) docker image with
our `terraform-provider-commercetools`.

To build the docker image file locally, use:
```sh
docker build . -t terraform-with-provider-commercetools:latest
```
Then you can run a terraform command on files in the current directory with:
```sh
docker run -v${pwd}:/config terraform-with-provider-commercetools:latest <CMD>
```

## Authors
This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-commercetools)
for more information.
