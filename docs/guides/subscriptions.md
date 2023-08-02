---
subcategory: ""
page_title: "Subscriptions"
description: |-
    Creating subscriptions in the various cloud providers
---

## AWS Subscription Example

```hcl
provider "aws" {
  region = "eu-west-1"
}

provider "commercetools" {
  client_id     = "foo"
  client_secret = "bar"
  project_key   = "some-project"
  scopes        = "manage_project:some-project"
  token_url     = "https://auth.sphere.io"
  api_url       = "https://api.sphere.io"
}

resource "aws_sqs_queue" "ct_queue" {
  name                      = "terraform-queue-two"
  delay_seconds             = 90
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_user" "ct" {
  name = "specific-user"
}

resource "aws_iam_access_key" "ct" {
  user = aws_iam_user.ct.name
}

resource "aws_iam_user_policy" "policy" {
  name = "commercetools-access"
  user = aws_iam_user.ct.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "sqs:sqs:SendMessage"
      ],
      "Effect": "Allow",
      "Resource": "${aws_sqs_queue.ct_queue.arn}"
    }
  ]
}
EOF
}

resource "commercetools_subscription" "subscribe" {
  key = "my-subscription"

  destination {
    type          = "SQS"
    queue_url     = aws_sqs_queue.ct_queue.id
    access_key    = aws_iam_access_key.ct.id
    access_secret = aws_iam_access_key.ct.secret
    region        = "eu-west-1"
  }

  changes {
    resource_type_ids = ["product"]
  }

  message {
    resource_type_id = "product"
    types            = ["ProductPublished", "ProductCreated"]
  }
}
```

## Google Pubsub Example

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

resource "google_pubsub_topic" "resource-updates" {
  name = "resource-updates"
}

# add ctp subscription service account
resource "google_pubsub_topic_iam_member" "ctp-subscription-publisher" {
  topic  = google_pubsub_topic.resource-updates.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:subscriptions@commercetools-platform.iam.gserviceaccount.com"
}

resource "commercetools_subscription" "subscribe" {
  key = "my-subscription"

  destination {
    type        = "google_pubsub"
    project_id  = var.project
    topic       = google_pubsub_topic.resource-updates.name
  }

  changes {
    resource_type_ids = [
      "category",
      "product",
      "product-type",
      "customer-group"
    ]
  }

  depends_on = [
    "google_pubsub_topic_iam_member.ctp-subscription-publisher"
  ]
}
```

## Azure EventGrid Example

```hcl
provider "azurerm" {} #initiate properly

provider "commercetools" {
  client_id     = "foo"
  client_secret = "bar"
  project_key   = "some-project"
  scopes        = "manage_project:some-project"
  token_url     = "https://auth.sphere.io"
  api_url       = "https://api.sphere.io"
}

resource "azurerm_eventgrid_topic" "order_changes" {
  name                = "my_topic_name"
  location            = var.azure_resource_group.location
  resource_group_name = var.azure_resource_group.name
  input_schema        = "CloudEventSchemaV1_0"
}

resource "commercetools_subscription" "order_changes" {
  key = "commercetools_order_changes"

  destination {
    type       = "EventGrid"
    uri        = azurerm_eventgrid_topic.order_changes.endpoint
    access_key = azurerm_eventgrid_topic.order_changes.primary_access_key
  }

  changes {
    resource_type_ids = ["order"]
  }

  message {
    resource_type_id = "order"
    types            = ["OrderCreated", "OrderPaymentStateChanged"]
  }

  format {
    type                 = "CloudEvents"
    cloud_events_version = "1.0"
  }
}
```
