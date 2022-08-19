
resource "commercetools_state" "product_for_sale" {
  key  = "product-for-sale"
  type = "ProductState"
  name = {
    en = "For Sale"
  }
  description = {
    en = "Regularly stocked product."
  }
  initial = true
}

resource "commercetools_state" "product_clearance" {
  key  = "product-clearance"
  type = "ProductState"
  name = {
    en = "On Clearance"
  }
  description = {
    en = "The product line will not be ordered again."
  }
}


// Only allow transition from sale to clearance
resource "commercetools_state_transitions" "transition_1" {
  from = commercetools_state.product_for_sale.id
  to = [
    commercetools_state.product_clearance.id,
  ]
}

// Disable transitions from product clearance to other
resource "commercetools_state_transitions" "transition_2" {
  from = commercetools_state.product_clearance.id
  to   = []
}
