# terraform-provider-commercetools

[![Travis Build Status](https://travis-ci.org/labd/terraform-provider-commercetools.svg?branch=master)](https://travis-ci.org/labd/terraform-provider-commercetools)
[![codecov](https://codecov.io/gh/LabD/terraform-provider-commercetools/branch/master/graph/badge.svg)](https://codecov.io/gh/LabD/terraform-provider-commercetools)

Note: This is currently **NOT** ready for production usage

## Using the SDK

## Example

```hcl
provider "aws" {
  region = "eu-west-1"
}

provider "commercetools" {
  client_id     = "<your client id>"
  client_secret = "<your client secret>"
  project_key   = "<your project key>"
}

resource "aws_sqs_queue" "ct_queue" {
  name                      = "terraform-queue-two"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_user" "sqs_user" {
  name = "specific-user"
}

resource "aws_iam_access_key" "ct" {
  user = "${aws_iam_user.ct.name}"
}

resource "aws_iam_user_policy" "policy" {
  name = "commercetools-access"
  user = "${aws_iam_user.ct.name}"

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
    queue_url     = "${aws_sqs_queue.ct_queue.id}"
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

# Testing

## Running an Acceptance Test

Acceptance tests can be run using the `testacc` target in the
`Makefile`. Prior to running the tests provider configuration details such as
access keys must be made available as environment variables.

For example, to run an acceptance test against the CommerceTools API
provider, the following environment variables must be set:

```sh
export COMMERCETOOLS_CLIENT_ID=...
export COMMERCETOOLS_CLIENT_SECRET=...
export COMMERCETOOLS_PROJECT_KEY=...
```

For convenience, place a `testenv.sh` in your project root (which is included in .gitignore) where you can store these environment variables

Tests can then be run by running

```sh
  $ make testacc
```