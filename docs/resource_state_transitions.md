# State Transitions

State transitions allow you to define a subset of valid changes of state.
By default, if no state transitions are defined, a state can be transitioned to any other state.
Using state transitions narrows the number of states an initial state can be transitioned to.
Use a `state_transitions` with an empty list of transitions to prevent any transitioning of that state.

Also see the [states HTTP API documentation][commercetool-states].

## Example Usage

Product state transitions from on sale to on clearance or recalled.

```hcl
resource "commercetools_state" "product_for_sale" {
  key = "product-for-sale"
  type = "ProductState"
  name = {
    en = "For Sale"
  }
  initial = true
}

resource "commercetools_state" "product_clearance" {
  key = "product-clearance"
  type = "ProductState"
  name = {
    en = "On Clearance"
  }
}

resource "commercetools_state" "product_recalled" {
  key = "product-clearance"
  type = "ProductState"
  name = {
    en = "Recalled"
  }
}

resource "commercetools_state_transitions" "product_for_sale" {
  from = commercetools_state.product_for_sale.id
  to   = [
    commercetools_state.product_clearance.id,
    commercetools_state.product_recalled.id,
  ]
}
```

Defining a state that can not be transitioned to any other state.

```
resource "commercetools_state" "review_archived" {
  key = "review-archived"
  type = "ReviewState"
  name = {
    en = "Archived"
  }
  description = {
    en = "No longer viewable by customers."
  }
}

resource "commercetools_state_transitions" "review_archived" {
  from = commercetools_state.review_archived.id
  to   = []
}
```

## Argument Reference

The following arguments are supported:

* `from` - The id of the state to transition from.
* `to` - The ids of the states the "from" state can be transitioned to.

[commercetool-states]: https://docs.commercetools.com/http-api-projects-states.html

## Import

State transitions can be imported using the state id, e.g.

```
$ terraform import commercetools_state_transitions.example 838411e7-84a6-4553-b546-51fb29705529 
```