---
subcategory: ""
page_title: "Examples"
description: |-
    Example usage
---

# Examples

A few examples for different cloud providers.

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

## Google Pubsub Example

```hcl

provider "google" {
  project = "${var.project}"
  credentials = "${file("${var.credentials_file_path}")}"
}

resource "google_pubsub_topic" "resource-updates" {
  name = "resource-updates"
}

# add ctp subscription service account
resource "google_pubsub_topic_iam_member" "ctp-subscription-publisher" {
  topic = "${google_pubsub_topic.resource-updates.name}"
  role = "roles/pubsub.publisher"
  member = "serviceAccount:subscriptions@commercetools-platform.iam.gserviceaccount.com"
}

provider "commercetools" {
}

resource "commercetools_subscription" "subscribe" {
  key = "my-subscription"

  destination {
    type = "google_pubsub"
    project_id = "${var.project}"
    topic = "${google_pubsub_topic.resource-updates.name}"
  }

  changes {
    resource_type_ids = [
      "category",
      "product",
      "product-type",
      "customer-group"
    ]
  }
  depends_on = [ "google_pubsub_topic_iam_member.ctp-subscription-publisher" ]
}
```


## Types example

```hcl
resource "commercetools_type" "ctype1" {
  key = "contact_info"
  name = {
    en = "Contact info"
    nl = "Contact informatie"
  }
  description = {
    en = "All things related communication"
    nl = "Alle communicatie-gerelateerde zaken"
  }

  resource_type_ids = ["customer"]
  
  field {
    name = "skype_name"
    label = {
      en = "Skype name"
      nl = "Skype naam"
    }
    type {
      name = "String"
    }
  }

  field {
    name = "contact_time"
    label = {
      en = "Contact time"
      nl = "Contact tijd"
    }
    type {
      name = "Enum"
      values {
        day = "Daytime"
        evening = "Evening"
      }
    }
  }

  field {
    name = "contact_preference"
    label = {
      en = "Contact preference"
      nl = "Contact voorkeur"
    }
    type {
      name = "LocalizedEnum"
      localized_value {
        key = "phone"
        label {
          en = "Phone"
          nl = "Telefoon"
        }
      }
      localized_value {
        key = "skype"
        label {
          en = "Skype"
          nl = "Skype"
        }
      }
    }
  }
}
```
