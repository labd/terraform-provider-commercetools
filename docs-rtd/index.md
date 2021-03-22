# Commercetools Terraform provider
This is the Terraform provider for commercetools. It allows you to configure your
commercetools project with infrastructure-as-code principles.


# Commercial support
Need support implementing this terraform module in your organization? We are
able to offer support. Please contact us at
[opensource@labdigital.nl](opensource@labdigital.nl)!


## Installation
Terraform 0.13 added support for automatically downloading providers from
the terraform registry. Add the following to your terraform project

```hcl
terraform {
  required_providers {
    commercetools = {
      source = "labd/commercetools"
    }
  }
}
```

Packages of the releases are available at [the GitHub Repo](https://github.com/labd/terraform-provider-commercetools/releases).
See the [terraform documentation](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins)
for more information about installing third-party providers.


## Using the provider
The provider attempts to read the required values from environment variables:
- `CTP_CLIENT_ID`
- `CTP_CLIENT_SECRET`
- `CTP_PROJECT_KEY`
- `CTP_SCOPES`
- `CTP_API_URL`
- `CTP_AUTH_URL`

Alternatively, you can set it up directly in the terraform file:

```hcl
provider "commercetools" {
  client_id     = "<your client id>"
  client_secret = "<your client secret>"
  project_key   = "<your project key>"
  project_key   = "<your project key>"
  scopes        = "<space seperated list of scopes>"
  api_url       = "<api url>"
  token_url     = "<token url>"
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
