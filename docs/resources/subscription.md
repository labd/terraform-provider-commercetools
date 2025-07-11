---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commercetools_subscription Resource - terraform-provider-commercetools"
subcategory: ""
description: |-
  Subscriptions allow you to be notified of new messages or changes via a Message Queue of your choice. Subscriptions are used to trigger an asynchronous background process in response to an event on the commercetools platform. Common use cases include sending an Order Confirmation Email, charging a Credit Card after the delivery has been made, or synchronizing customer accounts to a Customer Relationship Management (CRM) system.
  See also the Subscriptions API Documentation https://docs.commercetools.com/api/projects/subscriptions
---

# commercetools_subscription (Resource)

Subscriptions allow you to be notified of new messages or changes via a Message Queue of your choice. Subscriptions are used to trigger an asynchronous background process in response to an event on the commercetools platform. Common use cases include sending an Order Confirmation Email, charging a Credit Card after the delivery has been made, or synchronizing customer accounts to a Customer Relationship Management (CRM) system.

See also the [Subscriptions API Documentation](https://docs.commercetools.com/api/projects/subscriptions)

## Example Usage

```terraform
resource "commercetools_subscription" "my-sqs-subscription" {
  key = "my-sqs-subscription-key"
  destination {
    type          = "SQS"
    queue_url     = aws_sqs_queue.your-queue.id
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

  event {
    resource_type_id = "import-api"
    types            = ["ImportContainerCreated"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `changes` (Block Set) The change notifications subscribed to (see [below for nested schema](#nestedblock--changes))
- `destination` (Block List) (see [below for nested schema](#nestedblock--destination))
- `event` (Block Set) Events to be subscribed to (see [below for nested schema](#nestedblock--event))
- `format` (Block List) The [format](https://docs.commercetools.com/api/projects/subscriptions#format) in which the payload is delivered (see [below for nested schema](#nestedblock--format))
- `key` (String) Timestamp of the last Terraform update of the order.
- `message` (Block Set) The messages subscribed to (see [below for nested schema](#nestedblock--message))

### Read-Only

- `id` (String) The ID of this resource.
- `version` (Number)

<a id="nestedblock--changes"></a>
### Nested Schema for `changes`

Required:

- `resource_type_ids` (List of String) [Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#changesubscription)


<a id="nestedblock--destination"></a>
### Nested Schema for `destination`

Required:

- `type` (String) The type of the destination. See [Destination](https://docs.commercetools.com/api/projects/subscriptions#destination) for more information

Optional:

- `access_key` (String, Sensitive) The access key of the SQS queue, SNS topic or EventBridge topic
- `access_secret` (String, Sensitive) The access secret of the SQS queue, SNS topic or EventBridge topic
- `account_id` (String) The AWS account ID of the SNS topic or EventBridge topic
- `acks` (String) The acks value of the Confluent Cloud topic
- `api_key` (String) The API key of the Confluent Cloud topic
- `api_secret` (String) The API secret of the Confluent Cloud topic
- `bootstrap_server` (String) The bootstrap server of the Confluent Cloud topic
- `connection_string` (String) The connection string of the Azure Service Bus
- `key` (String) The key of the Confluent Cloud topic
- `project_id` (String) The project ID of the Google Cloud Pub/Sub
- `queue_url` (String) The URL of the SQS queue
- `region` (String) The region of the SQS queue, SNS topic or EventBridge topic
- `topic` (String) The topic of the Google Cloud Pub/Sub or Confluent Cloud topic
- `topic_arn` (String) The ARN of the SNS topic
- `uri` (String) The URI of the EventGrid topic


<a id="nestedblock--event"></a>
### Nested Schema for `event`

Required:

- `resource_type_id` (String) [Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#ctp:api:type:EventSubscriptionResourceTypeId)
- `types` (List of String) Must contain valid event types for the resource. For example, for resource type import-api the event type ImportContainerCreated is valid. If no types are given, the Subscription will receive all events for the defined resource type.


<a id="nestedblock--format"></a>
### Nested Schema for `format`

Optional:

- `cloud_events_version` (String) For CloudEvents
- `type` (String)


<a id="nestedblock--message"></a>
### Nested Schema for `message`

Required:

- `resource_type_id` (String) [Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#changesubscription)
- `types` (List of String) types must contain valid message types for this resource, for example for resource type product the message type ProductPublished is valid. If no types of messages are given, the subscription is valid for all messages of this resource
