resource "commercetools_state" "review_unreviewed" {
  key  = "review-unreviewed"
  type = "ReviewState"
  name = {
    en = "Unreviewed"
  }
  description = {
    en = "Not reviewed yet"
  }
  initial = true
  roles   = ["ReviewIncludedInStatistics"]
}

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
