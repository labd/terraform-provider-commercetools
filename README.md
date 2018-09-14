# commercetools Terraform Provider

[![Travis Build Status](https://travis-ci.org/labd/terraform-provider-commercetools.svg?branch=master)](https://travis-ci.org/labd/terraform-provider-commercetools)
[![codecov](https://codecov.io/gh/LabD/terraform-provider-commercetools/branch/master/graph/badge.svg)](https://codecov.io/gh/LabD/terraform-provider-commercetools)
[![Go Report Card](https://goreportcard.com/badge/github.com/labd/terraform-provider-commercetools)](https://goreportcard.com/report/github.com/labd/terraform-provider-commercetools)

Note: This is currently **NOT** ready for production usage

## Requirements

## Using the provider

### Example

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

resource "aws_iam_user" "ct" {
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
    resource_type_id = ["product"]
  }

  message {
    resource_type_id = "product"
    types            = ["ProductPublished", "ProductCreated"]
  }
}
```

# Contributing

## Building the provider

Clone repository to: `$GOPATH/src/github.com/labd/terraform-provider-commercetools`

```sh
$ mkdir -p $GOPATH/src/github.com/labd; cd $GOPATH/src/github.com/labd
$ git clone git@github.com:labd/terraform-provider-commercetools
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/labd/terraform-provider-commercetools
$ make build
```

## Testing

### Running the unit tests

```sh
$ make test
```

### Running an Acceptance Test

In order to run the full suite of Acceptance tests, run `make testacc`.

**NOTE:** Acceptance tests create real resources.

Prior to running the tests provider configuration details such as access keys
must be made available as environment variables.

Since we need to be able to create commercetools resources, we need the
commercetools API credentials. So in order for the acceptance tests to run
correctly please provide all of the following:

```sh
export CTP_CLIENT_ID=...
export CTP_CLIENT_SECRET=...
export CTP_PROJECT_KEY=...
```
For convenience, place a `testenv.sh` in your `local` folder (which is
included in .gitignore) where you can store these environment variables.

Tests can then be started by running

```sh
$ source local/testenv.sh
$ make testacc
```
