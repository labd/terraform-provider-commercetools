# The Initial state is a state that is provided in a new commercetools environment by default
data "commercetools_state" "initial_state" {
  key = "Initial"
}

resource "commercetools_state_transitions" "from_created_to_allocated" {
  from = data.commercetools_state.initial_state.id
  to = [
    commercetools_state.backorder.id,
  ]
}

resource "commercetools_state" "backorder" {
  key  = "backorder"
  type = "LineItemState"
  name = {
    en = "Back Order",

  }
  description = {
    en = "Not available - on back order"
  }
  initial = false
}
