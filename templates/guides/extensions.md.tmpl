---
subcategory: ""
page_title: "API Extensions"
description: |-
    API Extensions on Google Cloud Functions, AWS Lambda or any HTTP endpoint
---

# API Extensions

An [API Extension](https://docs.commercetools.com/api/projects/api-extensions)
is called by commercetools while a resource such as a cart, order, payment,
customer, business unit, quote, shopping list or product is created or
updated, and can change it or reject the request. It is set up with `commercetools_api_extension`,
which points at where the extension runs: a Google Cloud Function, an AWS
Lambda function, or any HTTP endpoint.

An extension is called on every matching request, within a time limit of 2
seconds (10 seconds for payments), so keep it fast and narrow its `trigger`
with a `condition` where possible.

The provider configuration is the same for each example:

```hcl
provider "commercetools" {
  client_id     = var.ctp_client_id
  client_secret = var.ctp_client_secret
  project_key   = "my-project"
  scopes        = "manage_project:my-project"
  api_url       = "https://api.europe-west1.gcp.commercetools.com"
  token_url     = "https://auth.europe-west1.gcp.commercetools.com"
}
```

## Google Cloud Functions

With the `GoogleCloudFunction` destination, commercetools calls the function as
the service account `extensions@commercetools-platform.iam.gserviceaccount.com`,
so the function needs no public access. A 2nd gen function runs on Cloud Run,
so the service account gets the Cloud Run Invoker role on it.

```hcl
locals {
  project = "my-gcp-project"
  region  = "europe-west1"
}

provider "google" {
  project = local.project
  region  = local.region
}

data "archive_file" "source" {
  type        = "zip"
  source_dir  = "src"
  output_path = "function-source.zip"
}

resource "google_storage_bucket" "source" {
  name                        = "${local.project}-gcf-source"
  location                    = "EU"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_object" "source" {
  name   = "function-source-${data.archive_file.source.output_md5}.zip"
  bucket = google_storage_bucket.source.name
  source = data.archive_file.source.output_path
}

resource "google_cloudfunctions2_function" "cart_extension" {
  name     = "cart-extension"
  location = local.region

  build_config {
    runtime     = "nodejs22"
    entry_point = "cartHandler"
    source {
      storage_source {
        bucket = google_storage_bucket.source.name
        object = google_storage_bucket_object.source.name
      }
    }
  }

  service_config {
    max_instance_count = 10
    available_memory   = "256M"
    timeout_seconds    = 10
  }
}

resource "google_cloud_run_service_iam_member" "commercetools_invoker" {
  location = google_cloudfunctions2_function.cart_extension.location
  service  = google_cloudfunctions2_function.cart_extension.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:extensions@commercetools-platform.iam.gserviceaccount.com"
}

resource "commercetools_api_extension" "cart" {
  key           = "cart-extension"
  timeout_in_ms = 2000

  destination {
    type = "GoogleCloudFunction"
    url  = google_cloudfunctions2_function.cart_extension.service_config[0].uri
  }

  trigger {
    resource_type_id = "cart"
    actions          = ["Create", "Update"]
  }

  depends_on = [google_cloud_run_service_iam_member.commercetools_invoker]
}
```

## AWS Lambda

commercetools invokes the function with the access key of an IAM user that may
only invoke it.

```hcl
provider "aws" {
  region = "eu-west-1"
}

resource "aws_iam_user" "commercetools" {
  name = "commercetools-cart-extension"
}

resource "aws_iam_access_key" "commercetools" {
  user = aws_iam_user.commercetools.name
}

resource "aws_iam_user_policy" "invoke" {
  name = "invoke-cart-extension"
  user = aws_iam_user.commercetools.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "lambda:InvokeFunction"
      Resource = aws_lambda_function.cart_extension.arn
    }]
  })
}

resource "commercetools_api_extension" "cart" {
  key = "cart-extension"

  destination {
    type          = "AWSLambda"
    arn           = aws_lambda_function.cart_extension.arn
    access_key    = aws_iam_access_key.commercetools.id
    access_secret = aws_iam_access_key.commercetools.secret
  }

  trigger {
    resource_type_id = "cart"
    actions          = ["Create", "Update"]
  }
}
```

## HTTP

Any HTTPS endpoint can be an extension. Protect it with a secret that
commercetools sends in the `Authorization` header, or, for Azure Functions,
with the function key in `azure_authentication`.

```hcl
variable "extension_secret" {
  type      = string
  sensitive = true
}

resource "commercetools_api_extension" "order" {
  key = "order-validation"

  destination {
    type                 = "HTTP"
    url                  = "https://extensions.example.com/order"
    authorization_header = "Bearer ${var.extension_secret}"
  }

  trigger {
    resource_type_id = "order"
    actions          = ["Create"]
    # Only orders shipped to the Netherlands.
    condition = "country = \"NL\""
  }
}
```
