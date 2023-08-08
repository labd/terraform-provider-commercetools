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

resource "google_cloudfunctions_function" "my_cloud_function" {
  name        = "function-test"
  description = "My function"
  runtime     = "nodejs16"

  # See https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function for any
  # further settings
}

resource "google_cloudfunctions_function_iam_member" "invoker" {
  # For GoogleCloudFunction destinations, you need to grant permissions to the
  # <extensions@commercetools-platform.iam.gserviceaccount.com> service account to invoke your function.
  project        = "my-project"
  region         = "europe-central2"
  cloud_function = google_cloudfunctions_function.my_cloud_function.name

  # If your function's version is 1st gen, grant the service account the IAM role Cloud Functions Invoker
  role = "roles/cloudfunctions.invoker"
  # For version 2nd gen, assign the IAM role Cloud Run Invoker
  # role   = "roles/run.invoker"
  member = "serviceAccount:extensions@commercetools-platform.iam.gserviceaccount.com"
}
