# Custom Application (Merchant Center)

Manage the resources for [Custom Applications](https://docs.commercetools.com/custom-applications).

## Example Usage

Assuming you are building a Custom Application to manage State Machines, you can provide the following configuration:

```hcl
resource "commercetools_custom_application" "state-machines" {
  name        = "State machines"
  description = "Manage state machines"
  url         = "https://state-machines.my-domain.com"
  is_active   = true
  navbar_menu {
    uri_path = "state-machines"
    icon     = "RocketIcon"
    label_all_locales {
      locale = "en"
      value  = "State machines"
    }
    label_all_locales {
      locale = "de"
      value  = "Zustandsmachinen"
    }
    permissions = ["ViewDeveloperSettings"]

    submenu {
      uri_path = "state-machines/new"
      label_all_locales {
        locale = "en"
        value  = "Add state machine"
      }
      label_all_locales {
        locale = "de"
        value  = "Zustandsmachine hinzufügen"
      }
      permissions = ["ManageDeveloperSettings"]
    }
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
