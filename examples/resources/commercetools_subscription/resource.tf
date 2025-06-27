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
