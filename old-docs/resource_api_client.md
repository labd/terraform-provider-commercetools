# API Client

Create a new API client. Note that Commercetools might return slightly different scopes, resulting in a new API client being created everytime Terraform is run. In this case,
fix your scopes accordingly to match what is returned by Commercetools.

Also see the [API client HTTP API documentation](https://docs.commercetools.com//http-api-projects-api-clients).

## Example Usage

```hcl
resource "commercetools_api_client" "my-api-client" {
  name = "My API Client"
  scope = ["manage_orders:my-ct-project-key", "manage_payments:my-ct-project-key"]
}

```

## Argument Reference

The following arguments are supported:

* `name` - Name of the API client
* `scope` - A list of the [OAuth scopes](https://docs.commercetools.com/http-api-authorization.html#scopes)
