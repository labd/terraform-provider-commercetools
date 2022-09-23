---
subcategory: ""
page_title: "API Extensions"
description: |-
    Using AWS Lambda, Google Cloud Functions for API Extensions
---

## Example

```hcl
locals {
  project = "<your project id>"
  region  = "europe-west1"
}

provider "commercetools" {
  client_id     = "foo"
  client_secret = "bar"
  project_key   = "some-project"
  scopes        = "manage_project:some-project"
  token_url     = "https://auth.sphere.io"
  api_url       = "https://api.sphere.io"
}

provider "google" {
  project = local.project
  region  = local.region
}


# Create the artifact
data "archive_file" "source" {
  type        = "zip"
  source_dir  = "src"
  output_path = "functions-source.zip"
}

resource "google_storage_bucket" "bucket" {
  name                        = "${local.project}-gcf-source"
  location                    = "EU"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_object" "object" {
  name   = "function-source-${data.archive_file.source.output_md5}.zip"
  bucket = google_storage_bucket.bucket.name
  source = "function-source.zip"
}

resource "google_cloudfunctions_function" "function" {
  name        = "my-cart-extension"
  region      = local.region
  description = "Extension for cart create / update"

  runtime      = "nodejs16"
  trigger_http = true
  entry_point  = "cartHandler"

  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.object.name
}


# Allow everyone to access this function. Commercetools doesn't support auth
# yet for cloud functions, so this is needed.
resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}

resource "commercetools_api_extension" "cart_extension" {
  key = "my-cart-extension"

  destination {
    type = "HTTP"
    url  = google_cloudfunctions_function.function.https_trigger_url
  }

  trigger {
    resource_type_id = "cart"
    actions          = ["Create", "Update"]
  }
}
```
