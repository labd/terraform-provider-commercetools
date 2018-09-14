# Subscription

Provides an commercetools subscription

## Example Usage

```hcl
resource "commercetools_subscription" "my-sqs-subscription" {
  key = "my-subscription"

  destination {
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



#### Destination
A destination contains all info necessary for the commercetools platform to
deliver a message onto your Message Queue. Message Queues can be
differentiated by the type field.
