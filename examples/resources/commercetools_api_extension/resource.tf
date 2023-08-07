# HTTP api extension
resource "commercetools_api_extension" "my-http-extension" {
  key = "my-http-extension-key"

  destination {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "customer"
    actions          = ["Create", "Update"]
  }
}

# AWS Lambda api extension
resource "commercetools_api_extension" "my-awslambda-extension" {
  key = "my-awslambda-extension-key"

  destination {
    type          = "awslambda"
    arn           = "us-east-1:123456789012:mylambda"
    access_key    = "mykey"
    access_secret = "mysecret"
  }

  trigger {
    resource_type_id = "customer"
    actions          = ["Create", "Update"]
  }
}

# Google Cloud Function api extension
resource "commercetools_api_extension" "my-googlecloudfunction-extension" {
  key = "my-googlecloudfunction-extension-key"

  destination {
    type = "googlecloudfunction"
    url  = "https://example.com"
  }

  trigger {
    resource_type_id = "customer"
    actions          = ["Create", "Update"]
  }
}

