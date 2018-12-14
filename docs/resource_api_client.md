# API Client

Create a new API client.

Also see the [API client HTTP API documentation](https://docs.commercetools.com//http-api-projects-api-clients).

## Example Usage

```hcl
resource "commercetools_api_client" "my-api-client" {
  name = "My API Client"
  scope = "manage_orders:my-ct-project-key manage_payments:my-ct-project-key"
}

```

## Argument Reference

The following arguments are supported:

* `name` - Name of the API client
* `scope` - A whitespace separated list of the [OAuth scopes](https://docs.commercetools.com/http-api-authorization.html#scopes) 
