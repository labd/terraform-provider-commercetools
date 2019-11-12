# Subscription

Provides a commercetools subscription

Also see the [subscription HTTP API documentation](https://docs.commercetools.com/http-api-projects-subscriptions.html).

## Example Usage

```hcl
resource "commercetools_subscription" "my-sqs-subscription" {
  key = "my-subscription"

  destination = {
    type          = "SQS"
    queue_url     = "${aws_sqs_queue.your-queue.id}"
    access_key    = "${aws_iam_access_key.ct.id}"
    access_secret = "${aws_iam_access_key.ct.secret}"
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

## Argument Reference

The following arguments are supported:

* `key` - User-specific unique identifier for the subscription
* `destination` - The [Message Queue](#destination) into which the notifications are to be sent
* `changes` - The change notifications subscribed to.
* `messages` - The messages subscribed to.
* `format` - The format in which the payload is delivered.

### Destination

A destination contains all info necessary for the commercetools platform to
deliver a message onto your Message Queue. Message Queues can be
differentiated by the type field.

#### AWS SQS Destination

* `type` - `"SQS"`
* `queue_url` - The url of the queue.
* `access_key` - The aws access key.
* `access_secret` - The aws access secret.
* `region` - The aws region.

#### AWS SNS Destination

* `type` - `"SNS"`
* `topic_arn` - The arn of the topic.
* `access_key` - The aws access key.
* `access_secret` - The aws access secret.

#### Azure Service Bus Destination

* `type` - `"azure_servicebus"`
* `connection_string` - The SharedAccessKey for the service bus destination.

#### Azure Event Grid Destination

* `type` - `"azure_eventgrid"`
* `uri` - The URI of the topic.
* `access_key` - The access key for the destination.

#### Google Cloud Pub/Sub Destination

* `type` - `"google_pubsub"`
* `project_id` - The id of the project that contains the Pub/Sub topic.
* `topic` - The name of the Pub/Sub topic.
