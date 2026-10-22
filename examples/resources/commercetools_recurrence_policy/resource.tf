resource "commercetools_recurrence_policy" "monthly" {
  key = "monthly"
  name = {
    en = "Every month"
  }
  schedule {
    type          = "standard"
    value         = 1
    interval_unit = "Months"
  }
}

resource "commercetools_recurrence_policy" "first_of_month" {
  key = "first-of-month"
  schedule {
    type = "dayOfMonth"
    day  = 1
  }
}
