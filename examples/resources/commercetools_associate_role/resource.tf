resource "commercetools_type" "my-type" {
  key = "my-type"
  name = {
    en = "My type"
    nl = "Mijn type"
  }

  resource_type_ids = ["associate-role"]

  field {
    name = "my-field"
    label = {
      en = "My field"
      nl = "Mijn veld"
    }
    type {
      name = "String"
    }
  }
}

resource "commercetools_associate_role" "my-role" {
  key              = "my-role"
  buyer_assignable = false
  name             = "My Role"
  permissions = [
    "AddChildUnits",
    "UpdateAssociates",
    "UpdateBusinessUnitDetails",
    "UpdateParentUnit",
    "ViewMyCarts",
    "ViewOthersCarts",
    "UpdateMyCarts",
    "UpdateOthersCarts",
    "CreateMyCarts",
    "CreateOthersCarts",
    "DeleteMyCarts",
    "DeleteOthersCarts",
    "ViewMyOrders",
    "ViewOthersOrders",
    "UpdateMyOrders",
    "UpdateOthersOrders",
    "CreateMyOrdersFromMyCarts",
    "CreateMyOrdersFromMyQuotes",
    "CreateOrdersFromOthersCarts",
    "CreateOrdersFromOthersQuotes",
    "ViewMyQuotes",
    "ViewOthersQuotes",
    "AcceptMyQuotes",
    "AcceptOthersQuotes",
    "DeclineMyQuotes",
    "DeclineOthersQuotes",
    "RenegotiateMyQuotes",
    "RenegotiateOthersQuotes",
    "ReassignMyQuotes",
    "ReassignOthersQuotes",
    "ViewMyQuoteRequests",
    "ViewOthersQuoteRequests",
    "UpdateMyQuoteRequests",
    "UpdateOthersQuoteRequests",
    "CreateMyQuoteRequestsFromMyCarts",
    "CreateQuoteRequestsFromOthersCarts",
    "CreateApprovalRules",
    "UpdateApprovalRules",
    "UpdateApprovalFlows",
  ]

  custom {
    type_id = commercetools_type.my-type.id
    fields = {
      my-field = "My value"
    }
  }
}
