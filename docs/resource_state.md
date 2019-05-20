# States

The commercetools platform allows you to model states of certain objects, such as orders, line items, products, reviews, and payments in order to define finite state machines reflecting the business logic you'd like to implement.

Also see the [states HTTP API documentation][commercetool-states].

## Example Usage

```hcl
resource "commercetools_state" "review_unreviewed" {
  key = "review-unreviewed"
  type = "ReviewState"
  name = {
      en = "Unreviewed"
  }
  description = {
    en = "Not reviewed yet"
  }
  initial = true
  roles = ["ReviewIncludedInStatistics"]
}
```

## Argument Reference

The following arguments are supported:

* `key` - A unique identifier for the state.
* `type` - Which CTP resource or object the state shall belong to. See [Commercetools documentation][commercetools-states] for possible values.
* `name` - Optional, localized name of the state.
* `description` - Optional, localized description of the state.
* `initial` - Optional, initial state of the state machine.
* `roles` - Optional, list of roles this state has. See [Commercetools documentation][commercetools-states] for possible values.

**Transitions are not yet supported**

[commercetool-states]: https://docs.commercetools.com/http-api-projects-states.html
