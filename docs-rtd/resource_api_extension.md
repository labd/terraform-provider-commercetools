# API Extension

Provides a commercetools API extension

Also see the [extension HTTP API documentation](https://docs.commercetools.com/http-api-projects-api-extensions).

## Example Usage

```hcl
resource "commercetools_api_extension" "my-extension" {
  key = "test-case"

  destination {
    type = "HTTP"
    url = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions          = ["Create", "Update"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `key` - User-specific unique identifier for the subscription
* `destination` - Details where the extension can be reached
* `triggers` - Describes what triggers the extension
* `timeout_in_ms` - The maximum time the commercetools platform waits for a response from the extension. If not present,
  2000 (2 seconds) is used.

### Azure Functions

The `destination` field supports Azure functions. In this case pass the `azure_authentication` value instead
of `authorization_header` like so:

```hcl
resource "commercetools_api_extension" "my-extension" {
  key = "test-case"

  destination = {
    type = "http"
    url = "https://some_azure_url"
    azure_authentication = "an_azure_function_key"
  }

  trigger {
    resource_type_id = "payment"
    actions = ["Create", "Update"]
  }

  timeout_in_ms = 2000
}
```

### AWS Lambda
The `destination` field supports AWS Lambda functions. In this case set `type` to be `awslambda` and pass 
`arn`, `access_key` and `access_secret` fields like so: 

```hcl
resource "commercetools_api_extension" "my-extension" {
  key = "test-case"

  destination = {
    type = "awslambda"
    arn = "arn:aws:lambda:some_lambda_arn"
    access_key = "access_key_123"
    access_secret = "123secretabc"
  }

  trigger {
    resource_type_id = "customer"
    actions = ["Create", "Update"]
  }

  timeout_in_ms = 2000
}
```
