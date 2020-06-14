# States

The commercetools platform allows you to model states of certain objects, such as orders, line items, products, reviews, and payments in order to define finite state machines reflecting the business logic you'd like to implement.

Also see the [states HTTP API documentation][commercetool-states].

## Example Usage

Review state:

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

State with transitions specified:

```hcl
resource "commercetools_state" "product_for_sale" {
  key = "product-for-sale"
  type = "ProductState"
  name = {
      en = "For Sale"
  }
  description = {
    en = "Regularly stocked product."
  }
  initial = true
}

resource "commercetools_state_transitions" "product_for_sale" {
  from = commercetools_state.product_for_sale.id
  to   = [commercetools_state.product_clearance.id]
}

resource "commercetools_state" "product_clearance" {
  key = "product-clearance"
  type = "ProductState"
  name = {
      en = "On Clearance"
  }
  description = {
    en = "The product line will not be ordered again."
  }
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

If you want to declare state transitions, use the [`commercetools_state_transitions`](/docs/resource_state_transitions.md) resource.

[commercetool-states]: https://docs.commercetools.com/http-api-projects-states.html
