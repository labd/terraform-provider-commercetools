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

# An API extension which runs after another extension, expands references in the
# payload and receives the state of the resource before the update
resource "commercetools_api_extension" "my-dependent-extension" {
  key = "my-dependent-extension-key"

  destination {
    type                 = "HTTP"
    url                  = "https://example.com"
    authorization_header = "Basic 12345"
  }

  trigger {
    resource_type_id = "cart"
    actions          = ["Create", "Update"]
  }

  # Wait for these extensions to complete before this one is called. A maximum
  # of 5 dependencies is allowed and the chain may not be deeper than 3 layers.
  # Every entry is either a key or an id; an entry formatted as a UUID is treated
  # as an id. Referencing by key also works for extensions that are not managed
  # by this Terraform configuration.
  dependencies = [
    commercetools_api_extension.my-http-extension.key,
    "an-extension-managed-elsewhere",
  ]

  # Reference expansion applied to both `resource` and `oldResource`
  expansion_paths = ["lineItems[*].variant", "customerGroup"]

  additional_context {
    # Include an `oldResource` field with the state before the update. Only
    # applies to `Update` actions.
    include_old_resource = true
  }
}
