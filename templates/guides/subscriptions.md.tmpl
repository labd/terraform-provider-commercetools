---
subcategory: ""
page_title: "Subscriptions"
description: |-
    Sending messages, changes and events to AWS, Google Cloud, Azure or Confluent Cloud
---

# Subscriptions

A [Subscription](https://docs.commercetools.com/api/projects/subscriptions)
sends notifications to a queue or topic in your cloud, from where a service
picks them up. It is set up with `commercetools_subscription`, with what it
subscribes to and a `destination`.

A subscription can subscribe to three kinds of notifications, in any
combination:

- `message`: the [Messages](https://docs.commercetools.com/api/projects/messages)
  of a resource type, such as `OrderCreated`. Messages have to be enabled on
  the Project, with `messages { enabled = true }` on `commercetools_project_settings`.
- `changes`: a notification each time a resource of a type is created,
  updated or deleted.
- `event`: the events of other APIs, such as the Import API's `ImportContainerCreated`.

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

## AWS SQS

With the access key left out, commercetools sends to the queue with IAM
authentication, as the AWS user `arn:aws:iam::362576667341:user/subscriptions`,
which the queue's policy allows to send. There are no keys to manage.

```hcl
provider "aws" {
  region = "eu-west-1"
}

resource "aws_sqs_queue" "orders" {
  name                      = "commercetools-orders"
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_sqs_queue_policy" "commercetools" {
  queue_url = aws_sqs_queue.orders.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "arn:aws:iam::362576667341:user/subscriptions" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.orders.arn
    }]
  })
}

resource "commercetools_subscription" "orders" {
  key = "orders"

  destination {
    type      = "SQS"
    queue_url = aws_sqs_queue.orders.id
    region    = "eu-west-1"
  }

  message {
    resource_type_id = "order"
    types            = ["OrderCreated", "OrderStateChanged"]
  }

  # commercetools sends a test message on creation, which needs the policy.
  depends_on = [aws_sqs_queue_policy.commercetools]
}
```

An SNS topic works the same way, with `type = "SNS"`, `topic_arn`, and a topic
policy that allows `sns:Publish`.

## AWS EventBridge

For an EventBridge destination, commercetools creates a partner event source
in the AWS account. Associate it with an event bus in AWS to complete the
setup.

```hcl
resource "commercetools_subscription" "products" {
  key = "products"

  destination {
    type       = "EventBridge"
    region     = "eu-west-1"
    account_id = "123456789012"
  }

  changes {
    resource_type_ids = ["product"]
  }
}
```

## Google Cloud Pub/Sub

The topic has to allow the service account
`subscriptions@commercetools-platform.iam.gserviceaccount.com` to publish.

```hcl
locals {
  project = "my-gcp-project"
}

provider "google" {
  project = local.project
  region  = "europe-west1"
}

resource "google_pubsub_topic" "resource_updates" {
  name = "commercetools-resource-updates"
}

resource "google_pubsub_topic_iam_member" "commercetools_publisher" {
  topic  = google_pubsub_topic.resource_updates.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:subscriptions@commercetools-platform.iam.gserviceaccount.com"
}

resource "commercetools_subscription" "resource_updates" {
  key = "resource-updates"

  destination {
    type       = "GoogleCloudPubSub"
    project_id = local.project
    topic      = google_pubsub_topic.resource_updates.name
  }

  changes {
    resource_type_ids = ["category", "product", "product-type", "customer-group"]
  }

  depends_on = [google_pubsub_topic_iam_member.commercetools_publisher]
}
```

## Azure Event Grid

```hcl
provider "azurerm" {
  features {}
}

resource "azurerm_eventgrid_topic" "order_changes" {
  name                = "commercetools-order-changes"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  input_schema        = "CloudEventSchemaV1_0"
}

resource "commercetools_subscription" "order_changes" {
  key = "order-changes"

  destination {
    type       = "EventGrid"
    uri        = azurerm_eventgrid_topic.order_changes.endpoint
    access_key = azurerm_eventgrid_topic.order_changes.primary_access_key
  }

  message {
    resource_type_id = "order"
    types            = ["OrderCreated", "OrderPaymentStateChanged"]
  }

  # Event Grid's topic expects CloudEvents.
  format {
    type                 = "CloudEvents"
    cloud_events_version = "1.0"
  }
}
```

## Azure Service Bus

The queue or topic needs a shared access policy with the `Send` permission,
whose connection string goes into the destination.

```hcl
resource "azurerm_servicebus_queue" "orders" {
  name         = "commercetools-orders"
  namespace_id = azurerm_servicebus_namespace.main.id
}

resource "azurerm_servicebus_queue_authorization_rule" "commercetools" {
  name     = "commercetools"
  queue_id = azurerm_servicebus_queue.orders.id
  send     = true
}

resource "commercetools_subscription" "orders" {
  key = "orders"

  destination {
    type              = "AzureServiceBus"
    connection_string = azurerm_servicebus_queue_authorization_rule.commercetools.primary_connection_string
  }

  message {
    resource_type_id = "order"
    types            = ["OrderCreated"]
  }
}
```

## Confluent Cloud

```hcl
resource "commercetools_subscription" "orders" {
  key = "orders"

  destination {
    type             = "ConfluentCloud"
    bootstrap_server = "pkc-12345.europe-west1.gcp.confluent.cloud:9092"
    api_key          = var.confluent_api_key
    api_secret       = var.confluent_api_secret
    topic            = "commercetools-orders"
    acks             = "all"
  }

  message {
    resource_type_id = "order"
    types            = ["OrderCreated"]
  }
}
```
